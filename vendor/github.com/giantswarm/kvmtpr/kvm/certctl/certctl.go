package certctl

import (
	"github.com/giantswarm/kvmtpr/kvm/certctl/docker"
)

type Certctl struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
