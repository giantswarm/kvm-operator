# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.13.0] - 2020-10-30

### Changed

- Update `k8scloudconfig` to v7.2.0, containing a fix for DockerHub QPS.

## [3.12.2] - 2020-10-16

### Added

- Add monitoring labels

## [3.12.1] - 2020-07-28

### Changed

- Stop provisioning NGINX ingress controller NodePort Service.

## [3.12.0] - 2020-07-09

### Added

- Add compatibility for Kubernetes 1.17.
- Use tags for images, templated into ignition.

### Changed

- Improved upgrades from earlier KVM v11.X releases.
- Fix cluster creation with `float64` wrapper in CRD.
- Use `k8s-kvm:0.2.0` with QEMU 4.2.0.

## [3.11.1] 2020-04-30

### Changed

- Use Release.Revision in Helm chart for Helm 3 support.
- Fix OIDC settings.


## [3.11.0] 2020-04-27

### Added

- First release as a flattened operator.
- Support setting OIDC username and groups prefix.
- Add `conntrackMaxPerCore` parameter in kube-proxy manifest.

### Changed

- Use Flatcar linux instead of CoreOS.
- Streamlined image templating for core components for quicker and easier releases in the future.
- Retrieve component versions from `releases`.
- Remove debug profiling from Controller Manager and Scheduler
- Remove limit of calico node init container.

[Unreleased]: https://github.com/giantswarm/kvm-operator/compare/v3.13.0...HEAD
[3.13.0]: https://github.com/giantswarm/kvm-operator/compare/v3.12.2...v3.13.0
[3.12.2]: https://github.com/giantswarm/kvm-operator/compare/v3.12.1...v3.12.2
[3.12.1]: https://github.com/giantswarm/kvm-operator/compare/v3.12.0...v3.12.1
[3.12.0]: https://github.com/giantswarm/kvm-operator/compare/v3.11.1...v3.12.0
[3.11.1]: https://github.com/giantswarm/kvm-operator/compare/v3.11.0...v3.11.1
[3.11.0]: https://github.com/giantswarm/kvm-operator/releases/tag/v3.11.0
