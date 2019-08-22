package installation

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/dns"
	"github.com/giantswarm/kvm-operator/flag/service/installation/ntp"
	"github.com/giantswarm/kvm-operator/flag/service/installation/tenant"
)

type Installation struct {
	DNS    dns.DNS
	Name   string
	NTP    ntp.NTP
	Tenant tenant.Tenant
}
