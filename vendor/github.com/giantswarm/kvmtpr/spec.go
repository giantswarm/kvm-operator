package kvmtpr

import (
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/kvmtpr/kvm"
)

type Spec struct {
	Cluster clustertpr.Cluster `json:"cluster" yaml:"cluster"`
	KVM     kvm.KVM            `json:"kvm" yaml:"kvm"`
}
