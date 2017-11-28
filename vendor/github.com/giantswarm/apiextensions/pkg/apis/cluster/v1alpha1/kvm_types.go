package v1alpha1

import (
	"net"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewKVMCRD returns a new custom resource definition for KVM. This might look
// something like the following.
//
//     apiVersion: apiextensions.k8s.io/v1beta1
//     kind: CustomResourceDefinition
//     metadata:
//       name: kvms.cluster.giantswarm.io
//     spec:
//       group: cluster.giantswarm.io
//       scope: Namespaced
//       version: v1alpha1
//       names:
//         kind: KVM
//         plural: kvms
//         singular: kvm
//
func NewKVMCRD() *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiextensionsv1beta1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kvms.cluster.giantswarm.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "cluster.giantswarm.io",
			Scope:   "Namespaced",
			Version: "v1alpha1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:     "KVM",
				Plural:   "kvms",
				Singular: "kvm",
			},
		},
	}
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KVM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              KVMSpec `json:"spec"`
}

type KVMSpec struct {
	Cluster       KVMSpecCluster       `json:"cluster" yaml:"cluster"`
	KVM           KVMSpecKVM           `json:"kvm" yaml:"kvm"`
	VersionBundle KVMSpecVersionBundle `json:"versionBundle" yaml:"versionBundle"`
}

type KVMSpecCluster struct {
	Calico     KVMSpecClusterCalico     `json:"calico" yaml:"calico"`
	Customer   KVMSpecClusterCustomer   `json:"customer" yaml:"customer"`
	Docker     KVMSpecClusterDocker     `json:"docker" yaml:"docker"`
	Etcd       KVMSpecClusterEtcd       `json:"etcd" yaml:"etcd"`
	ID         string                   `json:"id" yaml:"id"`
	Kubernetes KVMSpecClusterKubernetes `json:"kubernetes" yaml:"kubernetes"`
	Masters    []KVMSpecClusterNode     `json:"masters" yaml:"masters"`
	Vault      KVMSpecClusterVault      `json:"vault" yaml:"vault"`
	Workers    []KVMSpecClusterNode     `json:"workers" yaml:"workers"`
}

type KVMSpecClusterCalico struct {
	CIDR   int    `json:"cidr" yaml:"cidr"`
	Domain string `json:"domain" yaml:"domain"`
	MTU    int    `json:"mtu" yaml:"mtu"`
	Subnet string `json:"subnet" yaml:"subnet"`
}

type KVMSpecClusterCustomer struct {
	ID string `json:"id" yaml:"id"`
}

type KVMSpecClusterDocker struct {
	Daemon KVMSpecClusterDockerDaemon `json:"daemon" yaml:"daemon"`
}

type KVMSpecClusterDockerDaemon struct {
	CIDR      string `json:"cidr" yaml:"cidr"`
	ExtraArgs string `json:"extraArgs" yaml:"extraArgs"`
}

type KVMSpecClusterEtcd struct {
	AltNames string `json:"altNames" yaml:"altNames"`
	Domain   string `json:"domain" yaml:"domain"`
	Port     int    `json:"port" yaml:"port"`
	Prefix   string `json:"prefix" yaml:"prefix"`
}

type KVMSpecClusterKubernetes struct {
	API               KVMSpecClusterKubernetesAPI               `json:"api" yaml:"api"`
	DNS               KVMSpecClusterKubernetesDNS               `json:"dns" yaml:"dns"`
	Domain            string                                    `json:"domain" yaml:"domain"`
	Hyperkube         KVMSpecClusterKubernetesHyperkube         `json:"hyperkube" yaml:"hyperkube"`
	IngressController KVMSpecClusterKubernetesIngressController `json:"ingressController" yaml:"ingressController"`
	Kubelet           KVMSpecClusterKubernetesKubelet           `json:"kubelet" yaml:"kubelet"`
	NetworkSetup      KVMSpecClusterKubernetesNetworkSetup      `json:"networkSetup" yaml:"networkSetup"`
	SSH               KVMSpecClusterKubernetesSSH               `json:"ssh" yaml:"ssh"`
}

type KVMSpecClusterKubernetesAPI struct {
	AltNames       string `json:"altNames" yaml:"altNames"`
	ClusterIPRange string `json:"clusterIPRange" yaml:"clusterIPRange"`
	Domain         string `json:"domain" yaml:"domain"`
	IP             net.IP `json:"ip" yaml:"ip"`
	InsecurePort   int    `json:"insecurePort" yaml:"insecurePort"`
	SecurePort     int    `json:"securePort" yaml:"securePort"`
}

type KVMSpecClusterKubernetesDNS struct {
	IP net.IP `json:"ip" yaml:"ip"`
}

type KVMSpecClusterKubernetesHyperkube struct {
	Docker KVMSpecClusterKubernetesHyperkubeDocker `json:"docker" yaml:"docker"`
}

type KVMSpecClusterKubernetesHyperkubeDocker struct {
	Image string `json:"image" yaml:"image"`
}

type KVMSpecClusterKubernetesIngressController struct {
	Docker         KVMSpecClusterKubernetesIngressControllerDocker `json:"docker" yaml:"docker"`
	Domain         string                                          `json:"domain" yaml:"domain"`
	WildcardDomain string                                          `json:"wildcardDomain" yaml:"wildcardDomain"`
	InsecurePort   int                                             `json:"insecurePort" yaml:"insecurePort"`
	SecurePort     int                                             `json:"securePort" yaml:"securePort"`
}

type KVMSpecClusterKubernetesIngressControllerDocker struct {
	Image string `json:"image" yaml:"image"`
}

type KVMSpecClusterKubernetesKubelet struct {
	AltNames string `json:"altNames" yaml:"altNames"`
	Domain   string `json:"domain" yaml:"domain"`
	Labels   string `json:"labels" yaml:"labels"`
	Port     int    `json:"port" yaml:"port"`
}

type KVMSpecClusterKubernetesNetworkSetup struct {
	Docker KVMSpecClusterKubernetesNetworkSetupDocker `json:"docker" yaml:"docker"`
}

type KVMSpecClusterKubernetesNetworkSetupDocker struct {
	Image string `json:"image" yaml:"image"`
}

type KVMSpecClusterKubernetesSSH struct {
	UserList []KVMSpecClusterKubernetesSSHUser `json:"userList" yaml:"userList"`
}

type KVMSpecClusterKubernetesSSHUser struct {
	Name      string `json:"name" yaml:"name"`
	PublicKey string `json:"publicKey" yaml:"publicKey"`
}

type KVMSpecClusterNode struct {
	ID string `json:"id" yaml:"id"`
}

type KVMSpecClusterVault struct {
	Address string `json:"address" yaml:"address"`
	Token   string `json:"token" yaml:"token"`
}

type KVMSpecKVM struct {
	EndpointUpdater KVMSpecKVMEndpointUpdater `json:"endpointUpdater" yaml:"endpointUpdater"`
	K8sKVM          KVMSpecKVMK8sKVM          `json:"k8sKVM" yaml:"k8sKVM"`
	Masters         []KVMSpecKVMNode          `json:"masters" yaml:"masters"`
	NodeController  KVMSpecKVMNodeController  `json:"nodeController" yaml:"nodeController"`
	Workers         []KVMSpecKVMNode          `json:"workers" yaml:"workers"`
}

type KVMSpecKVMEndpointUpdater struct {
	Docker KVMSpecKVMEndpointUpdaterDocker `json:"docker" yaml:"docker"`
}

type KVMSpecKVMEndpointUpdaterDocker struct {
	Image string `json:"image" yaml:"image"`
}

type KVMSpecKVMK8sKVM struct {
	Docker      KVMSpecKVMK8sKVMDocker `json:"docker" yaml:"docker"`
	StorageType string                 `json:"storageType" yaml:"storageType"`
}

type KVMSpecKVMK8sKVMDocker struct {
	Image string `json:"image" yaml:"image"`
}

type KVMSpecKVMNode struct {
	CPUs   int     `json:"cpus" yaml:"cpus"`
	Disk   float64 `json:"disk" yaml:"disk"`
	Memory string  `json:"memory" yaml:"memory"`
}

type KVMSpecKVMNodeController struct {
	Docker KVMSpecKVMNodeControllerDocker `json:"docker" yaml:"docker"`
}

type KVMSpecKVMNodeControllerDocker struct {
	Image string `json:"image" yaml:"image"`
}

type KVMSpecVersionBundle struct {
	Version string `json:"version" yaml:"version"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KVMList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []KVM `json:"items"`
}
