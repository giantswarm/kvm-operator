package v1alpha1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewIngressCRD returns a new custom resource definition for Ingress. This
// might look something like the following.
//
//     apiVersion: apiextensions.k8s.io/v1beta1
//     kind: CustomResourceDefinition
//     metadata:
//       name: ingresss.core.giantswarm.io
//     spec:
//       group: core.giantswarm.io
//       scope: Namespaced
//       version: v1alpha1
//       names:
//         kind: Ingress
//         plural: ingresss
//         singular: ingress
//
func NewIngressCRD() *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiextensionsv1beta1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ingresss.core.giantswarm.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "core.giantswarm.io",
			Scope:   "Namespaced",
			Version: "v1alpha1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:     "Ingress",
				Plural:   "ingresss",
				Singular: "ingress",
			},
		},
	}
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Ingress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              IngressSpec `json:"spec"`
}

type IngressSpec struct {
	GuestCluster  IngressSpecGuestCluster   `json:"guestcluster" yaml:"guestcluster"`
	HostCluster   IngressSpecHostCluster    `json:"hostcluster" yaml:"hostcluster"`
	ProtocolPorts []IngressSpecProtocolPort `json:"protocolPorts" yaml:"protocolPorts"`
}

type IngressSpecGuestCluster struct {
	ID        string `json:"id" yaml:"id"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Service   string `json:"service" yaml:"service"`
}

type IngressSpecHostCluster struct {
	IngressController IngressSpecHostClusterIngressController `json:"ingressController" yaml:"ingressController"`
}

type IngressSpecHostClusterIngressController struct {
	ConfigMap string `json:"configMap" yaml:"configMap"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Service   string `json:"service" yaml:"service"`
}

type IngressSpecProtocolPort struct {
	IngressPort int    `json:"ingressPort" yaml:"ingressPort"`
	LBPort      int    `json:"lbPort" yaml:"lbPort"`
	Protocol    string `json:"protocol" yaml:"protocol"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type IngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Ingress `json:"items"`
}
