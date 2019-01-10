package installation

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/dns"
	"github.com/giantswarm/kvm-operator/flag/service/installation/guest"
)

type Installation struct {
	DNS   dns.DNS
	Guest guest.Guest
	Name  string
}
