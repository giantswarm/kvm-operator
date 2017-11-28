package v1alpha1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCertCRD returns a new custom resource definition for Cert. This might look
// something like the following.
//
//     apiVersion: apiextensions.k8s.io/v1beta1
//     kind: CustomResourceDefinition
//     metadata:
//       name: certs.core.giantswarm.io
//     spec:
//       group: core.giantswarm.io
//       scope: Namespaced
//       version: v1alpha1
//       names:
//         kind: Cert
//         plural: certs
//         singular: cert
//
func NewCertCRD() *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiextensionsv1beta1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "certs.core.giantswarm.io",
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "core.giantswarm.io",
			Scope:   "Namespaced",
			Version: "v1alpha1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:     "Cert",
				Plural:   "certs",
				Singular: "cert",
			},
		},
	}
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Cert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              CertSpec `json:"spec"`
}

type CertSpec struct {
	Cert          CertSpecCert          `json:"cert" yaml:"cert"`
	VersionBundle CertSpecVersionBundle `json:"versionBundle" yaml:"versionBundle"`
}

type CertSpecCert struct {
	AllowBareDomains bool     `json:"allowBareDomains" yaml:"allowBareDomains"`
	AltNames         []string `json:"altNames" yaml:"altNames"`
	ClusterComponent string   `json:"clusterComponent" yaml:"clusterComponent"`
	ClusterID        string   `json:"clusterID" yaml:"clusterID"`
	CommonName       string   `json:"commonName" yaml:"commonName"`
	IPSANs           []string `json:"ipSans" yaml:"ipSans"`
	Organizations    []string `json:"organizations" yaml:"organizations"`
	TTL              string   `json:"ttl" yaml:"ttl"`
}

type CertSpecVersionBundle struct {
	Version string `json:"version" yaml:"version"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Cert `json:"items"`
}
