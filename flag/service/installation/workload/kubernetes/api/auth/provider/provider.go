package provider

import (
	"github.com/giantswarm/kvm-operator/v4/flag/service/installation/workload/kubernetes/api/auth/provider/oidc"
)

type Provider struct {
	OIDC oidc.OIDC
}
