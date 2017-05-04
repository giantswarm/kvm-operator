package iptables

import (
	"github.com/giantswarm/kvmtpr/kvm/network/iptables/docker"
)

type IPTables struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
