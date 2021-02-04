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

### Pre-commit Hooks

`kvm-operator` uses [pre-commit](https://pre-commit.com/) to ensure that only good commits are pushed to the git
repository. It will have no effect unless `pre-commit` hooks have been installed after cloning the repository on your
develpoment machine. First, ensure that it is installed with `pip install pre-commit` or `brew install pre-commit`
(macOS). Then, install the git hooks in the root of the kvm-operator directory with `pre-commit install`. Any future
`git commit`s will automatically run the automated checks which include the following:

- `end-of-file-fixer`: Adds a final newline to any files missing one.
- `trailing-whitespace`: Removes trailing whitespace from all committed files.
- `no-commit-to-branch`: Prevents committing to `master`, `main`, and `release-*` branches.
- `check-merge-conflict`: Ensures that no merge conflict markers are found in source files.
- `go-test-repo-mod`: Ensures that all tests pass (`go test ./...`).
- `go-imports`: Ensures that imports are correctly sorted.
- `golangci-lint`: Ensure that `golangci-lint run` finds no problems.
- `go-build`: Ensures that `go build` returns no errors.
- `go-mod-tidy`: Ensures that `go mod tidy` doesn't change `go.sum`.

Note: `goimports` and `golangci-lint` should be available in your `$PATH` for these to run.

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
