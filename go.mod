module github.com/giantswarm/kvm-operator/v4

go 1.16

require (
	github.com/giantswarm/apiextensions/v6 v6.6.0
	github.com/giantswarm/badnodedetector/v2 v2.0.0
	github.com/giantswarm/certs/v4 v4.0.0
	github.com/giantswarm/errors v0.3.0
	github.com/giantswarm/k8sclient/v7 v7.0.1
	github.com/giantswarm/k8scloudconfig/v17 v17.0.0
	github.com/giantswarm/microendpoint v1.0.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v1.0.0
	github.com/giantswarm/micrologger v1.0.0
	github.com/giantswarm/operatorkit/v8 v8.0.0
	github.com/giantswarm/randomkeys/v3 v3.0.0
	github.com/giantswarm/statusresource/v5 v5.0.0
	github.com/giantswarm/tenantcluster/v6 v6.0.0
	github.com/giantswarm/to v0.3.0
	github.com/giantswarm/versionbundle v1.0.0
	github.com/google/go-cmp v0.5.6
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/viper v1.8.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
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

	// Fix [CVE-2020-26160] jwt-go before 4.0.0-preview1 allows attackers to bypass intended access restrict...
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1

	// Use v1.3.2 of gogo/protobuf to fix nancy alert.
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2

	// keep in sync with giantswarm/apiextensions
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v1.4.2
)
