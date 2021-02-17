package auth

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/workload/kubernetes/api/auth/provider"
)

type Auth struct {
	Provider provider.Provider
}
