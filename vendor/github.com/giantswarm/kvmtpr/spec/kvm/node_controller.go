package kvm

import "github.com/giantswarm/kvmtpr/spec/kvm/nodecontroller"

type NodeController struct {
	Docker nodecontroller.Docker `json:"docker" yaml:"docker"`
}
