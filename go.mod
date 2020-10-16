module github.com/giantswarm/kvm-operator/v3

go 1.14

require (
	github.com/giantswarm/apiextensions v0.4.20
	github.com/giantswarm/certs v0.2.0
	github.com/giantswarm/errors v0.2.3
	github.com/giantswarm/k8sclient v0.2.0
	github.com/giantswarm/k8scloudconfig/v7 v7.1.2
	github.com/giantswarm/kvm-operator v0.0.0-20201015133802-2dd74202625a
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.3.3
	github.com/giantswarm/operatorkit v0.2.1
	github.com/giantswarm/randomkeys v0.2.0
	github.com/giantswarm/statusresource v0.4.0
	github.com/giantswarm/tenantcluster v0.2.0
	github.com/giantswarm/versionbundle v0.2.0
	github.com/google/go-cmp v0.5.2
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/viper v1.7.1
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
)

// v3.3.17 is required by sigs.k8s.io/controller-runtime v0.5.2. Can remove this replace when updated.
replace github.com/coreos/etcd v3.3.17+incompatible => github.com/coreos/etcd v3.3.24+incompatible
