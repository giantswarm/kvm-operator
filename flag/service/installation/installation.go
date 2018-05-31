package installation

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/guest"
)

type Installation struct {
	Name  string
	Guest guest.Guest
}
