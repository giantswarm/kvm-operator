package config

import (
	"github.com/giantswarm/kvmtpr/kvm/network/config/docker"
)

type Config struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
