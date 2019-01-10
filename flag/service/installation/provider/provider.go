package provider

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/provider/kvm"
)

type Provider struct {
	KVM kvm.KVM
}
