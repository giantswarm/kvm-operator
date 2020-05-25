# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Compatibility with kubernetes 1.17.

### Changed

- Fix cluster creation with `float64` wrapper in CRD.


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

[Unreleased]: https://github.com/giantswarm/kvm-operator/compare/v3.11.1...HEAD
[3.11.1]: https://github.com/giantswarm/kvm-operator/releases/tag/v3.11.1
[3.11.0]: https://github.com/giantswarm/kvm-operator/releases/tag/v3.11.0
