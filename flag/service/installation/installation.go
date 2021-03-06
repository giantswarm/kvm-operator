package installation

import (
	"github.com/giantswarm/kvm-operator/v4/flag/service/installation/dns"
	"github.com/giantswarm/kvm-operator/v4/flag/service/installation/ntp"
	"github.com/giantswarm/kvm-operator/v4/flag/service/installation/workload"
)

type Installation struct {
	DNS      dns.DNS
	Name     string
	NTP      ntp.NTP
	Workload workload.Workload
}
