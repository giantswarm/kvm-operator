module github.com/giantswarm/kvm-operator

go 1.14

require (
	github.com/giantswarm/apiextensions/v3 v3.13.0
	github.com/giantswarm/certs/v3 v3.1.0
	github.com/giantswarm/errors v0.2.3
	github.com/giantswarm/k8sclient/v5 v5.0.0
	github.com/giantswarm/k8scloudconfig/v10 v10.0.1-0.20210105182409-15ccf899d9e1
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.4.0
	github.com/giantswarm/operatorkit/v4 v4.1.0
	github.com/giantswarm/randomkeys/v2 v2.0.0
	github.com/giantswarm/statusresource/v3 v3.0.0
	github.com/giantswarm/tenantcluster/v4 v4.0.0
	github.com/giantswarm/versionbundle v0.2.0
	github.com/google/go-cmp v0.5.4
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
)

// v3.3.10 is required by spf13/viper. Can remove this replace when updated.
replace github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible

// v3.3.13 is required by bketelsen/crypt. Can remove this replace when updated.
replace github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible

// keep in sync with giantswarm/apiextensions
replace sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.10-gs
