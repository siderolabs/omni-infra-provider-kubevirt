# Omni Infrastructure Provider for KubeVirt

Can be used to automatically provision Talos nodes in a KubeVirt cluster.

## Running Infrastructure Provider

First you need to create a service account for the infrastructure provider.

```bash
$ omnictl serviceaccount create --role=InfraProvider kubevirt

Set the following environment variables to use the service account:
OMNI_ENDPOINT=https://<account-name>.omni.siderolabs.io/
OMNI_SERVICE_ACCOUNT_KEY=<service-account-key>

Note: Store the service account key securely, it will not be displayed again
```

Create a service account kubeconfig for your KubeVirt cluster.
Store it in `kubeconfig` file.

### Using Docker

```bash
docker run -it -d -v ./kubeconfig:/kubeconfig ghcr.io/siderolabs/omni-infra-provider-kubevirt --kubeconfig /kubeconfig --kubeconfig kubeconfig --omni-api-endpoint https://<account-name>.omni.siderolabs.io/ --key <service-account-key>
```

### Using Executable

Build the project (should have docker and buildx installed):

```bash
make omni-infra-provider-linux-amd64
```

Run the executable:

```bash
_out/omni-infra-provider-linux-amd64 --kubeconfig kubeconfig --omni-api-endpoint https://<account-name>.omni.siderolabs.io/ --key <service-account-key>
```
