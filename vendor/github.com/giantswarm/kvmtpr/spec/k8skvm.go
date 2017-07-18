package spec

import "github.com/giantswarm/kvmtpr/spec/k8skvm"

type K8sKVM struct {
	Docker k8skvm.Docker `json:"docker" yaml:"docker"`
}
