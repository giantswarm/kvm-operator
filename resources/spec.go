package resources

import (
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
)

type ClusterList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []*Cluster `json:"items"`
}

type Cluster struct {
	unversioned.TypeMeta `json:",inline"`
	v1.ObjectMeta        `json:"metadata,omitempty"`
	Spec                 ClusterSpec `json:"spec"`
}

type ClusterSpec struct {
	Customer       string `json:"customer"`
	ClusterID      string `json:"clusterId"`
	WorkerReplicas int32  `json:"workerReplicas,omitempty"`

	K8sVmVersion           string `json:"k8sVmVersion"`
	FlannelClientVersion   string `json:"flannelClientVersion"`
	K8sVersion             string `json:"k8sVersion"`
	K8sNetworkSetupVersion string `json:"k8sNetworkSetupVersion"`

	ClusterVNI     int32  `json:"clusterVNI,omitempty"`
	ClusterBackend string `json:"clusterBackend"`
	ClusterNetwork string `json:"clusterNetwork"`
	CalicoSubnet   string `json:"calicoSubner"`
	CalicoCidr     string `json:"calicoCidr"`
	K8sCalicoMtu   string `json:"k8sCalicoMtu"`

	DockerExtraArgs string `json:"dockerExtraArgs,omitempty"`
	Registry        string `json:"registry"`

	MachineMem      string `json:"machineMem"`
	MachineCPUcores int32  `json:"machineCPUcores"`

	VaultToken string `json:"vaulToken"`

	K8sAPIaltNames       string `json:"k8sAPIaltNames"`
	K8sDomain            string `json:"k8sDomain"`
	K8sETCDdomainName    string `json:"k8sETCDdomainName"`
	K8sMasterDomainName  string `json:"k8sMasterDomainName"`
	K8sMasterServiceName string `json:"k8sMasterServiceName"`
	K8sNodeLabels        string `json:"k8sNodeLabels"`
	K8sClusterIpRange    string `json:"k8sClusterIpRange"`
	K8sClusterIpSubnet   string `json:"k8sClusterIpSubnet"`
	K8sMasterPort        string `json:"k8sMasterPort"`
	K8sDnsIp             string `json:"k8sDnsIp"`
	K8sInsecurePort      string `json:"k8sInsecurePort"`
	K8sSecurePort        string `json:"k8sSecurePort"`
	K8sWorkerServicePort string `json:"k8sWorkerServicePort"`

	// Ingress Controller
	KempVsIp              string `json:"kempVsIp"`
	KempVsName            string `json:"kempVsName"`
	KempVsPorts           string `json:"kempVsPorts"`
	KempVsSslAcceleration string `json:"kempVsSslAcceleration"`
	KempRsPort            string `json:"kempRsPort"`
	KempVsCheckPort       string `json:"kempVsCheckPort"`
	CloudflareIp          string `json:"cloudflareIp"`
	CloudflareDomain      string `json:"cloudflareDomain"`
}

func (c *Cluster) GetObjectKind() unversioned.ObjectKind {
	return nil
}
