module github.com/giantswarm/kvm-operator/v4

go 1.16

require (
	github.com/giantswarm/apiextensions/v3 v3.26.0
	github.com/giantswarm/badnodedetector v1.0.1
	github.com/giantswarm/certs/v3 v3.1.1
	github.com/giantswarm/errors v0.3.0
	github.com/giantswarm/k8sclient/v5 v5.11.0
	github.com/giantswarm/k8scloudconfig/v10 v10.8.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.5.0
	github.com/giantswarm/operatorkit/v5 v5.0.0
	github.com/giantswarm/randomkeys/v2 v2.1.0
	github.com/giantswarm/statusresource/v3 v3.1.0
	github.com/giantswarm/tenantcluster/v4 v4.1.0
	github.com/giantswarm/to v0.3.0
	github.com/giantswarm/versionbundle v0.2.0
	github.com/google/go-cmp v0.5.6
	github.com/prometheus/client_golang v1.10.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	k8s.io/api v0.18.19
	k8s.io/apimachinery v0.18.19
	k8s.io/client-go v0.18.19
	sigs.k8s.io/controller-runtime v0.6.4
)

replace (
	// v3.3.10 is required by spf13/viper. Can remove this replace when updated.
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible

	// v3.3.13 is required by bketelsen/crypt. Can remove this replace when updated.
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible

	// Use v1.3.2 of gogo/protobuf to fix nancy alert.
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2

	// keep in sync with giantswarm/apiextensions
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
)
