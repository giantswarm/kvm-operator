package kvm

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/provider/kvm/dns"
)

type KVM struct {
	DNS dns.DNS
}
