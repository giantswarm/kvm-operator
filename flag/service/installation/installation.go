package installation

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/guest"
	"github.com/giantswarm/kvm-operator/flag/service/installation/provider"
)

type Installation struct {
	Name     string
	Guest    guest.Guest
	Provider provider.Provider
}
