package spec

import "github.com/giantswarm/kvmtpr/spec/kvm"

type KVM struct {
	EndpointUpdater kvm.EndpointUpdater `json:"endpointUpdater" yaml:"endpointUpdater"`
	K8sKVM          kvm.K8sKVM          `json:"k8sKVM" yaml:"k8sKVM"`
	Masters         []kvm.Node          `json:"masters" yaml:"masters"`
	NodeController  kvm.NodeController  `json:"nodeController" yaml:"nodeController"`
	Workers         []kvm.Node          `json:"workers" yaml:"workers"`
}
