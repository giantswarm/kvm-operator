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
