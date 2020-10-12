module github.com/giantswarm/kvm-operator

go 1.14

require (
	github.com/giantswarm/apiextensions/v2 v2.6.0
	github.com/giantswarm/certs/v3 v3.1.0
	github.com/giantswarm/errors v0.2.3
	github.com/giantswarm/k8sclient/v4 v4.0.0
	github.com/giantswarm/k8scloudconfig/v8 v8.0.4-0.20201012201923-001c6733f699
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.3.3
	github.com/giantswarm/operatorkit/v2 v2.0.1
	github.com/giantswarm/randomkeys/v2 v2.0.0
	github.com/giantswarm/statusresource/v2 v2.0.0
	github.com/giantswarm/tenantcluster/v3 v3.0.0
	github.com/giantswarm/versionbundle v0.2.0
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/google/go-cmp v0.5.2
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	gopkg.in/ini.v1 v1.51.1 // indirect
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
)

// v3.3.X is required by sigs.k8s.io/controller-runtime. Can remove this replace when updated.
replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.25+incompatible
