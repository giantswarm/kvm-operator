package bridge

import (
	"github.com/giantswarm/kvmtpr/kvm/network/bridge/docker"
)

type Bridge struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
