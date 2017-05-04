package environment

import (
	"github.com/giantswarm/kvmtpr/kvm/network/environment/docker"
)

type Environment struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
