package k8skvm

import (
	"github.com/giantswarm/kvmtpr/kvm/k8skvm/docker"
)

type K8sKVM struct {
	Docker docker.Docker `json:"docker" yaml:"docker"`
}
