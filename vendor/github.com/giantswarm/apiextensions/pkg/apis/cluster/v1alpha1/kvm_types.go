package v1alpha1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewKvmConfigCRD returns a new custom resource definition for KvmConfig. This
// might look something like the following.
//
//     apiVersion: apiextensions.k8s.io/v1beta1
//     kind: CustomResourceDefinition
//     metadata:
//       name: kvmconfigs.cluster.giantswarm.io
//     spec:
//       group: cluster.giantswarm.io
//       scope: Namespaced
//       version: v1alpha1
//       names:
//         kind: KvmConfig
//         plural: kvmconfigs
//         singular: kvmconfig
//
func NewKvmConfigCRD() *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiextensionsv1beta1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kvmconfigs.cluster.giantswarm.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "cluster.giantswarm.io",
			Scope:   "Namespaced",
			Version: "v1alpha1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:     "KvmConfig",
				Plural:   "kvmconfigs",
				Singular: "kvmconfig",
			},
		},
	}
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KvmConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              KVMConfigSpec `json:"spec"`
}

type KVMConfigSpec struct {
	Cluster       Cluster                    `json:"cluster" yaml:"cluster"`
	KVM           KVMConfigSpecKVM           `json:"kvm" yaml:"kvm"`
	VersionBundle KVMConfigSpecVersionBundle `json:"versionBundle" yaml:"versionBundle"`
}

type KVMConfigSpecKVM struct {
	EndpointUpdater KVMConfigSpecKVMEndpointUpdater `json:"endpointUpdater" yaml:"endpointUpdater"`
	K8sKVM          KVMConfigSpecKVMK8sKVM          `json:"k8sKVM" yaml:"k8sKVM"`
	Masters         []KVMConfigSpecKVMNode          `json:"masters" yaml:"masters"`
	Network         KVMConfigSpecKVMNetwork         `json:"network" yaml:"network"`
	NodeController  KVMConfigSpecKVMNodeController  `json:"nodeController" yaml:"nodeController"`
	Workers         []KVMConfigSpecKVMNode          `json:"workers" yaml:"workers"`
}

type KVMConfigSpecKVMEndpointUpdater struct {
	Docker KVMConfigSpecKVMEndpointUpdaterDocker `json:"docker" yaml:"docker"`
}

type KVMConfigSpecKVMEndpointUpdaterDocker struct {
	Image string `json:"image" yaml:"image"`
}

type KVMConfigSpecKVMK8sKVM struct {
	Docker      KVMConfigSpecKVMK8sKVMDocker `json:"docker" yaml:"docker"`
	StorageType string                       `json:"storageType" yaml:"storageType"`
}

type KVMConfigSpecKVMK8sKVMDocker struct {
	Image string `json:"image" yaml:"image"`
}

type KVMConfigSpecKVMNode struct {
	CPUs   int     `json:"cpus" yaml:"cpus"`
	Disk   float64 `json:"disk" yaml:"disk"`
	Memory string  `json:"memory" yaml:"memory"`
}

type KVMConfigSpecKVMNetwork struct {
	Flannel KVMConfigSpecKVMNetworkFlannel `json:"flannel" yaml:"flannel"`
}

type KVMConfigSpecKVMNetworkFlannel struct {
	VNI int `json:"vni" yaml:"vni"`
}

type KVMConfigSpecKVMNodeController struct {
	Docker KVMConfigSpecKVMNodeControllerDocker `json:"docker" yaml:"docker"`
}

type KVMConfigSpecKVMNodeControllerDocker struct {
	Image string `json:"image" yaml:"image"`
}

type KVMConfigSpecVersionBundle struct {
	Version string `json:"version" yaml:"version"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KvmConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []KvmConfig `json:"items"`
}
