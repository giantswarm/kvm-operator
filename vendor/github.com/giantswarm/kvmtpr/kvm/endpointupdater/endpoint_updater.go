package endpointupdater

import (
	"github.com/giantswarm/kvmtpr/kvm/endpointupdater/docker"
)

type EndpointUpdater struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
