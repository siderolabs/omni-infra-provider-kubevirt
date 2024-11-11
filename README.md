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

If you are using `--data-volume-mode=Filesystem` (which is the default), make sure to enable the `ExpandDisks` featuregate in KubeVirt, e.g.:
```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
spec:
  configuration:
    developerConfiguration:
      featureGates:
        - ExpandDisks
```

By default VMs will use the bridge network binding mode. In IPv6 environments you might want to use [passt](https://kubevirt.io/user-guide/network/net_binding_plugins/passt/) instead. Make sure to set the provider configuration in your MachineClass accordingly.

### Using Docker

```bash
docker run -it -d -v ./kubeconfig:/kubeconfig ghcr.io/siderolabs/omni-infra-provider-kubevirt --kubeconfig-file /kubeconfig --omni-api-endpoint https://<account-name>.omni.siderolabs.io/ --omni-service-account-key <service-account-key> --data-volume-mode=Filesystem
```

### Using Executable

Build the project (should have docker and buildx installed):

```bash
make omni-infra-provider-linux-amd64
```

Run the executable:

```bash
_out/omni-infra-provider-linux-amd64 --kubeconfig kubeconfig-file --omni-api-endpoint https://<account-name>.omni.siderolabs.io/ --omni-service-account-key <service-account-key> --data-volume-mode=Filesystem
```
