package kvm

import (
	"github.com/giantswarm/kvmtpr/kvm/endpointupdater"
	"github.com/giantswarm/kvmtpr/kvm/k8skvm"
	"github.com/giantswarm/kvmtpr/kvm/kubectl"
	"github.com/giantswarm/kvmtpr/kvm/node"
)

type KVM struct {
	EndpointUpdater endpointupdater.EndpointUpdater `json:"endpointUpdater" yaml:"endpointUpdater"`
	K8sKVM          k8skvm.K8sKVM                   `json:"k8sKVM" yaml:"k8sKVM"`
	Kubectl         kubectl.Kubectl                 `json:"kubectl" yaml:"kubectl"`
	Masters         []node.Node                     `json:"masters" yaml:"masters"`
	Workers         []node.Node                     `json:"workers" yaml:"workers"`
}
