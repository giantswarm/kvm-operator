package spec

type KVM struct {
	EndpointUpdater EndpointUpdater `json:"endpointUpdater" yaml:"endpointUpdater"`
	K8sKVM          K8sKVM          `json:"k8sKVM" yaml:"k8sKVM"`
	Masters         []Node          `json:"masters" yaml:"masters"`
	Workers         []Node          `json:"workers" yaml:"workers"`
}
