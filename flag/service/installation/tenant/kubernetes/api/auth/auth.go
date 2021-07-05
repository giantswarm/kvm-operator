package auth

import (
	"github.com/giantswarm/kvm-operator/v4/flag/service/installation/tenant/kubernetes/api/auth/provider"
)

type Auth struct {
	Provider provider.Provider
}
