package provider

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/tenant/kubernetes/api/auth/provider/oidc"
)

type Provider struct {
	OIDC oidc.OIDC
}
