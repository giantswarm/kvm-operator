package kvm

import "github.com/giantswarm/kvmtpr/spec/kvm/network"

type Network struct {
	Flannel network.Flannel `json:"flannel" yaml:"flannel"`
}
