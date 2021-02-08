[![CircleCI](https://circleci.com/gh/giantswarm/kvm-operator.svg?&style=shield&circle-token=4434b93043ab299852583ebcd749440c9c700860)](https://circleci.com/gh/giantswarm/kvm-operator) [![Docker Repository on Quay](https://quay.io/repository/giantswarm/kvm-operator/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/kvm-operator)

# kvm-operator

`kvm-operator` manages Kubernetes clusters running on-premises in KVM VMs within a host Kubernetes cluster.


## Getting the Project

Download the latest release:
https://github.com/giantswarm/kvm-operator/releases/latest

Clone the git repository: https://github.com/giantswarm/kvm-operator.git

Download the latest docker image from here:
https://quay.io/repository/giantswarm/kvm-operator


### How to build

```
go build github.com/giantswarm/kvm-operator
```

## Architecture

The operator uses our [operatorkit][1] framework. It watches `KVMConfig`
CRs using a generated client stored in our [apiextensions][2] repo. Tenant clusters
each have a version known as a "tenant cluster version" which defines a tested set of
component versions such as Kubernetes and CoreDNS and are managed as Release CRs on the control plane.

The operator provisions tenant Kubernetes clusters running on-premises. It runs in a
host Kubernetes cluster running on bare metal or virtual machines.

[1]:https://github.com/giantswarm/operatorkit
[2]:https://github.com/giantswarm/apiextensions

### Controllers and Resource Handlers

kvm-operator contains four controllers each composed of one or more resource handlers:
- `cluster-controller` watches `KVMConfig`s and has the following handlers:
  - `clusterrolebinding`: Manages the RBAC and PSP role bindings used by WC node pods
  - `configmap`: Ensures that a configmap exists for each desired node containing the rendered ignition
  - `deployment`: Ensures that a deployment exists for each desired node
  - `ingress`: Manages the Kubernetes API and etcd ingresses for the WC
  - `namespace`: Manages the namespace for the cluster
  - `nodeindexstatus`: Manages node indexes in the KVMConfig status
  - `pvc`: Manages the PVC used to store WC etcd state (if PVC storage is enabled)
  - `service`: Manages the master and worker services
  - `serviceaccount`: Manages the service account used by WC node pods
  - `status`: Manages labels reflecting calculated status on WC objects
- `deleter-controller` watches `KVMConfig`s and has the following handlers:
  - `cleanupendpointips`: Ensures that worker and master Endpoints only contains IPs of Ready pods
  - `node`: Deletes `Node`s in the WC if they have no corresponding MC node pod
- `drainer-controller` watches `Pod`s and has the following handlers:
  - `endpoint`: Ensures that worker and master Endpoints exist and contain IPs of Ready pods only
  - `pod`: Prevents node pod deletion until draining of the corresponding WC node is complete
- `unhealthy-node-terminator-controller` watches `KVMConfig`s and has the following handlers:
  - `terminateunhealthynodes`: Deletes node pods when nodes are not ready for a certain period of time


### Kubernetes Resources

The operator creates a Kubernetes namespace per tenant cluster with a
service and endpoints. These are used by the control plane cluster to access the tenant
cluster.

### Certificates

Authentication for the cluster components and end-users uses TLS certificates.
These are provisioned using [Hashicorp Vault][5] and are managed by our
[cert-operator][6].

[5]:https://www.vaultproject.io/
[6]:https://github.com/giantswarm/cert-operator

### Endpoint management

Every `Service` object in Kubernetes generally has a corresponding `Endpoints` object with the same name, but if a
`Service` has no pod selectors, Kubernetes will not create an `Endpoints` object automatically. This allows the IPs a
`Service` points to be managed with a separate controller to point to any IPs (pod IPs or external IPs). `kvm-operator`
uses this approach to manage worker and master `Endpoints` objects for WCs.

`Endpoints` are managed by two different resource handlers, `drainer-controller`'s `endpoint` handler, and
`deleter-controller`'s `cleanupendpointips` handler. `deleter-controller` reconciles when events for a `KVMConfig` are
received whereas the `drainer-controller` updated when events for a `Pod` are received. This allows the operator to add
or remove endpoints when the cluster changes (such as during scaling) or when a pod changes (such as when draining an MC
node or when a pod becomes `NotReady`).


## Remote testing and debugging

Using `okteto`, we can synchronize local files (source files and compiled binaries) with a container in a remote
Kubernetes cluster to reduce the feedback loop when adding a feature or investigating a bug. Use the following commands
to get started.

#### Start okteto

- Download the latest Okteto release from https://github.com/okteto/okteto/releases.
- Ensure that `architect` is available in your `$PATH` (https://github.com/giantswarm/architect).
- Run `make okteto-up`.
- From this point on, you will be in a remote shell. You can modify files on your local machine, and they will be synced
  to the remote pod container.

#### Build and run operator inside remote container

- `go build`
- `./kvm-operator daemon --config.dirs=/var/run/kvm-operator/configmap/ --config.dirs=/var/run/kvm-operator/secret/ --config.files=config --config.files=dockerhub-secret`

#### Remote debugging with VS Code or Goland

- Install delve debugger in the remote container `cd /tmp && go get github.com/go-delve/delve/cmd/dlv && cd /okteto`
- Start delve server in the remote container `dlv debug --headless --listen=:2345 --log --api-version=2 -- daemon --config.dirs=/var/run/kvm-operator/configmap/ --config.dirs=/var/run/kvm-operator/secret/ --config.files=config --config.files=dockerhub-secret`.
- Wait until debug server is up (should see `API server listening at: [::]:2345`).
- Start the local debugger.
- If you make any changes to the source code, you will need to stop the debugging session, stop the server, and rebuild.

#### Clean up

- Run `make okteto-down`.

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- Bugs: [issues](https://github.com/giantswarm/kvm-operator/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

For security issues, please see [the security policy](SECURITY.md).


## License

kvm-operator is under the Apache 2.0 license. See the [LICENSE](LICENSE) file
for details.


## Credit
- https://golang.org
- https://github.com/giantswarm/microkit
