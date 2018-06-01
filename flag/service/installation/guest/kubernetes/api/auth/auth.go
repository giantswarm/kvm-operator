package auth

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/guest/kubernetes/api/auth/provider"
)

type Auth struct {
	Provider provider.Provider
}
