#!/bin/bash

set -eoux pipefail

TMP="/tmp/kubevirt-e2e"
mkdir -p "${TMP}"

# Settings.

TALOS_VERSION=1.9.4
OMNI_VERSION=${OMNI_VERSION:-latest}
K8S_VERSION="${K8S_VERSION:-v1.30.1}"

ARTIFACTS=_out
JOIN_TOKEN=testonly
RUN_DIR=$(pwd)
PLATFORM=$(uname -s | tr "[:upper:]" "[:lower:]")

export KUBECONFIG=${TMP}/kubeconfig

# Download required artifacts.

mkdir -p ${ARTIFACTS}

[ -f ${ARTIFACTS}/talosctl ] || (crane export ghcr.io/siderolabs/talosctl:latest | tar x -C ${ARTIFACTS})

# Schematic without any customizations
SCHEMATIC_ID="376567988ad370138ad8b2698212367b8edcb69b5fd68c80be1f2ec7d603b4ba"

TALOSCTL="${ARTIFACTS}/talosctl"
KUBECTL="${TMP}/kubectl"
OMNICTL="${TMP}/omnictl"

curl -Lo ${OMNICTL} $(curl https://api.github.com/repos/siderolabs/omni/releases/latest  |  jq -r '.assets[] | select(.name | contains ("omnictl-linux-amd64")) | .browser_download_url')
chmod +x ${OMNICTL}

curl -Lo ${KUBECTL} "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/${PLATFORM}/amd64/kubectl"
chmod +x ${KUBECTL}

# Build registry mirror args.

if [[ "${CI:-false}" == "true" ]]; then
  REGISTRY_MIRROR_FLAGS=()

  for registry in docker.io k8s.gcr.io quay.io gcr.io ghcr.io registry.k8s.io factory.talos.dev; do
    service="registry-${registry//./-}.ci.svc"
    addr=$(python3 -c "import socket; print(socket.gethostbyname('${service}'))")

    REGISTRY_MIRROR_FLAGS+=("--registry-mirror=${registry}=http://${addr}:5000")
  done
else
  # use the value from the environment, if present
  REGISTRY_MIRROR_FLAGS=("${REGISTRY_MIRROR_FLAGS:-}")
fi

CREATED_CLUSTER="kubevirt-test-$(echo $RANDOM | md5sum | head -c 10)"

function cleanup() {
  # gather container logs
  if [[ ! -z ${KUBECONFIG} ]]; then
    ${KUBECTL} get vm -A || true
    ${KUBECTL} get vmi -A || true
    ${KUBECTL} get datavolume -A || true
  fi

  if [[ "${CI:-false}" == "false" ]]; then
    rm -rf ${TMP}

    if [[ ! -z ${CREATED_CLUSTER} ]]; then
      echo "destroying created cluster"
      ${TALOSCTL} cluster destroy --name=${CREATED_CLUSTER} --provisioner=qemu || true
      rm -rf ~/.talos/clusters/${CREATED_CLUSTER}
    fi

    docker rm -f omni-integration vault-dev
  fi

  rm -rf _out/omni/
}

trap cleanup EXIT SIGINT

# Start Vault.

docker run --rm -d --cap-add=IPC_LOCK -p 8200:8200 -e 'VAULT_DEV_ROOT_TOKEN_ID=dev-o-token' --name vault-dev hashicorp/vault:1.15

sleep 10

# Load key into Vault.

docker cp hack/certs/key.private vault-dev:/tmp/key.private
docker exec -e VAULT_ADDR='http://0.0.0.0:8200' -e VAULT_TOKEN=dev-o-token vault-dev \
    vault kv put -mount=secret omni-private-key \
    private-key=@/tmp/key.private

sleep 5

# Launch Omni in the background.

export BASE_URL=https://localhost:8099/
export AUTH_USERNAME="${AUTH0_TEST_USERNAME}"
export AUTH0_CLIENT_ID="${AUTH0_CLIENT_ID}"
export AUTH0_DOMAIN="${AUTH0_DOMAIN}"

mkdir -p _out/omni/

docker run -it -d --network host -v ./hack/certs:/certs \
    -v $(pwd)/_out/omni:/_out \
    --cap-add=NET_ADMIN \
    --device=/dev/net/tun \
    -e SIDEROLINK_DEV_JOIN_TOKEN="${JOIN_TOKEN}" \
    -e VAULT_TOKEN=dev-o-token \
    -e VAULT_ADDR='http://127.0.0.1:8200' \
    --name omni \
    ghcr.io/siderolabs/omni:${OMNI_VERSION} \
    --siderolink-wireguard-advertised-addr 10.11.0.1:50180 \
    --siderolink-api-advertised-url "grpc://10.11.0.1:8090" \
    --machine-api-bind-addr 0.0.0.0:8090 \
    --siderolink-wireguard-bind-addr 0.0.0.0:50180 \
    --event-sink-port 8091 \
    --auth-auth0-enabled true \
    --advertised-api-url "${BASE_URL}" \
    --auth-auth0-client-id "${AUTH0_CLIENT_ID}" \
    --auth-auth0-domain "${AUTH0_DOMAIN}" \
    --initial-users "${AUTH_USERNAME}" \
    --private-key-source "vault://secret/omni-private-key" \
    --public-key-files "/certs/key.public" \
    --bind-addr 0.0.0.0:8099 \
    --key /certs/localhost-key.pem \
    --cert /certs/localhost.pem \
    --etcd-embedded-unsafe-fsync=true \
    --create-initial-service-account \
    --initial-service-account-key-path=/_out/key \
    "${REGISTRY_MIRROR_FLAGS[@]}"

docker logs -f omni &> ${TMP}/omni.log &

echo "creating cluster ${CREATED_CLUSTER}"
TAG="v${TALOS_VERSION}" ${TALOSCTL} cluster create \
  --name=${CREATED_CLUSTER} \
  --kubernetes-version=${K8S_VERSION} \
  ${REGISTRY_MIRROR_FLAGS} \
  --provisioner=qemu \
  --cidr 10.11.0.0/24 \
  --vmlinuz-path="https://factory.talos.dev/image/${SCHEMATIC_ID}/v${TALOS_VERSION}/kernel-amd64" \
  --initrd-path="https://factory.talos.dev/image/${SCHEMATIC_ID}/v${TALOS_VERSION}/initramfs-amd64.xz" \
  --controlplanes=3 \
  --workers=1 \
  --cpus-workers 16.0 \
  --memory 4096 \
  --memory-workers=32768 \
  --mtu 1430 \
  --config-patch @hack/test/configpatch.yaml \
  --disk=65536

${TALOSCTL} config nodes 10.11.0.2
${TALOSCTL} kubeconfig -f ${TMP}/kubeconfig

# install kubevirt

export RELEASE=$(curl https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)

${KUBECTL} apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-operator.yaml

${KUBECTL} apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-cr.yaml

${KUBECTL} apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/v1.60.1/cdi-operator.yaml
${KUBECTL} apply -f hack/test/manifests/cdi-cr.yaml
${KUBECTL} apply -f hack/test/manifests/local-path-storage.yaml

${KUBECTL} patch kubevirt -n kubevirt kubevirt --type='json' -p='[{"op": "replace", "path": "/spec/configuration", "value": {"developerConfiguration": {"featureGates": ["ExpandDisks"]}}}]'

${KUBECTL} -n kubevirt wait kv kubevirt --for condition=Available --timeout=10m

# Launch infra provider in the background

export OMNI_ENDPOINT=https://localhost:8099
export OMNI_SERVICE_ACCOUNT_KEY=$(cat _out/omni/key)

${OMNICTL} --insecure-skip-tls-verify infraprovider create kubevirt | tail -n5 | head -n2 | awk '{print "export " $0}' > ${TMP}/env

source ${TMP}/env

nice -n 10 ${ARTIFACTS}/omni-infra-provider-kubevirt-linux-amd64 \
  --kubeconfig-file=${KUBECONFIG} \
  --omni-api-endpoint https://localhost:8099 \
  --data-volume-mode Filesystem \
  --insecure-skip-verify&

docker run \
  -v $(pwd)/hack/certs:/etc/ssl/certs \
  -e SSL_CERT_DIR=/etc/ssl/certs \
  -e OMNI_SERVICE_ACCOUNT_KEY=$(cat _out/omni/key) \
  --network host \
  ghcr.io/siderolabs/omni-integration-test:${OMNI_VERSION} \
  --omni.endpoint https://localhost:8099 \
  --omni.talos-version=${TALOS_VERSION} \
  --test.run "TestIntegration/Suites/(ScaleUpAndDownAutoProvisionMachineSets)" \
  --omni.infra-provider=kubevirt \
  --omni.scale-timeout 5m \
  --omni.provider-data='{disk_size: 8, cores: 4, memory: 2048, architecture: amd64}' \
  --test.failfast \
  --test.v
