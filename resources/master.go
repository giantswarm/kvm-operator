package resources

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/giantswarm/clusterspec"

	"k8s.io/client-go/pkg/api"
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/intstr"
)

type Master interface {
	ClusterObj
}

type master struct {
	clusterspec.Cluster
}

func (m *master) generateMasterPodAffinity() (string, error) {
	podAffinity := &api.Affinity{
		PodAntiAffinity: &api.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []api.PodAffinityTerm{
				{
					LabelSelector: &apiunversioned.LabelSelector{
						MatchExpressions: []apiunversioned.LabelSelectorRequirement{
							{
								Key:      "app",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{"worker"},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
					Namespaces:  []string{m.Spec.ClusterId},
				},
			},
		},
		PodAffinity: &api.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []api.PodAffinityTerm{
				{
					LabelSelector: &apiunversioned.LabelSelector{
						MatchExpressions: []apiunversioned.LabelSelectorRequirement{
							{
								Key:      "app",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{"flannel-client"},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
					Namespaces:  []string{m.Spec.ClusterId},
				},
			},
		},
	}

	bytesPodAffinity, err := json.Marshal(podAffinity)
	if err != nil {
		return "", maskAny(err)
	}

	return string(bytesPodAffinity), nil
}

func (m *master) generateInitMasterContainers() (string, error) {
	privileged := true

	initContainers := []apiv1.Container{
		{
			Name:            "k8s-bridge-ip-configmap",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-bridge-ip-configmap:6d24a36be4d63259b67a1f46e3ff2d04a789e51c",
			ImagePullPolicy: apiv1.PullAlways,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "customer-dir",
					MountPath: "/tmp/",
				},
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "BRIDGE_IP_CONFIGMAP_NAME",
					Value: bridgeIPConfigmapName("master"),
				},
				{
					Name:  "BRIDGE_IP_CONFIGMAP_PATH",
					Value: bridgeIPConfigmapPath("master"),
				},
				{
					Name:  "K8S_NAMESPACE",
					Value: m.Spec.ClusterId,
				},
				{
					Name:  "NETWORK_BRIDGE_NAME",
					Value: networkBridgeName(m.Spec.ClusterId),
				},
			},
		},
		{
			Name: "kubectl-bridge-ip-configmap",
			// Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/kubectl:" + m.Spec.KubectlVersion,
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/kubectl:1.4.0",
			ImagePullPolicy: apiv1.PullAlways,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "customer-dir",
					MountPath: "/tmp/",
				},
			},
			Command: []string{
				"/bin/sh",
				"-c",
				"while [ ! -f ${BRIDGE_IP_CONFIGMAP_PATH} ]; do echo -; sleep 1; done; /usr/bin/kubectl --server=${G8S_MASTER_HOST}:${G8S_MASTER_PORT} replace --force -f ${BRIDGE_IP_CONFIGMAP_PATH}",
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "G8S_MASTER_PORT",
					Value: "8080",
				},
				{
					Name:  "G8S_MASTER_HOST",
					Value: "127.0.0.1",
				},
				{
					Name:  "BRIDGE_IP_CONFIGMAP_PATH",
					Value: bridgeIPConfigmapPath("master"),
				},
			},
		},
		{
			Name:            "k8s-master-api-token",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/alpine-openssl",
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/usr/bin/test ! -f /etc/kubernetes/secrets/token_sign_key.pem  && /usr/bin/openssl genrsa -out /etc/kubernetes/secrets/token_sign_key.pem 2048 && /bin/echo 'Generated new token sign key.' || /bin/echo 'Token sign key already exists, skipping.'",
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ssl",
					MountPath: "/etc/ssl/certs/ca-certificates.crt",
				},
				{
					Name:      "api-secrets",
					MountPath: "/etc/kubernetes/secrets",
				},
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
		},
		{
			Name:            "k8s-master-api-certs",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:" + m.Spec.CertctlVersion,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=api.$CLUSTER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/master/apiserver.pem --key-file=/etc/kubernetes/ssl/master/apiserver-key.pem --ca-file=/etc/kubernetes/ssl/master/apiserver-ca.pem --alt-names=$K8S_API_ALT_NAMES --ip-sans=$G8S_API_IP",
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ssl",
					MountPath: "/etc/ssl/certs/ca-certificates.crt",
				},
				{
					Name:      "api-certs",
					MountPath: "/etc/kubernetes/ssl/master/",
				},
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "K8S_MASTER_SERVICE_NAME",
					Value: m.Spec.Certificates.MasterServiceName,
				},
				{
					Name:  "K8S_API_ALT_NAMES",
					Value: m.Spec.Certificates.ApiAltNames,
				},
				{
					Name:  "G8S_API_IP",
					Value: m.Spec.GiantnetesConfiguration.ApiIp,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: m.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: m.Spec.ClusterId,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: m.Spec.Certificates.VaultToken,
				},
				{
					Name:  "VAULT_ADDR",
					Value: m.Spec.GiantnetesConfiguration.VaultAddr,
				},
			},
		},
		{
			Name:            "k8s-master-calico-certs",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:" + m.Spec.CertctlVersion,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=calico.$CLUSTER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/calico/client.pem --key-file=/etc/kubernetes/ssl/calico/client-key.pem --ca-file=/etc/kubernetes/ssl/calico/client-ca.pem --alt-names=$K8S_API_ALT_NAMES --ip-sans=$G8S_API_IP",
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ssl",
					MountPath: "/etc/ssl/certs/ca-certificates.crt",
				},
				{
					Name:      "calico-certs",
					MountPath: "/etc/kubernetes/ssl/calico/",
				},
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "K8S_MASTER_SERVICE_NAME",
					Value: m.Spec.Certificates.MasterServiceName,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: m.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: m.Spec.ClusterId,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: m.Spec.Certificates.VaultToken,
				},
				{
					Name:  "VAULT_ADDR",
					Value: m.Spec.GiantnetesConfiguration.VaultAddr,
				},
			},
		},
		{
			Name:            "k8s-master-etcd-certs",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:" + m.Spec.CertctlVersion,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=etcd.$CLUSTER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/etcd/server.pem --key-file=/etc/kubernetes/ssl/etcd/server-key.pem --ca-file=/etc/kubernetes/ssl/etcd/server-ca.pem --alt-names=$K8S_API_ALT_NAMES --ip-sans=$G8S_API_IP",
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ssl",
					MountPath: "/etc/ssl/certs/ca-certificates.crt",
				},
				{
					Name:      "etcd-certs",
					MountPath: "/etc/kubernetes/ssl/etcd/",
				},
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "K8S_MASTER_SERVICE_NAME",
					Value: m.Spec.Certificates.MasterServiceName,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: m.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: m.Spec.ClusterId,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: m.Spec.Certificates.VaultToken,
				},
				{
					Name:  "VAULT_ADDR",
					Value: m.Spec.GiantnetesConfiguration.VaultAddr,
				},
			},
		},
		{
			Name:            "set-iptables",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/alpine-bash-iptables",
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/sbin/iptables -I INPUT -p tcp --match multiport --dports $ETCD_PORT -d ${NODE_IP} -i ${NETWORK_BRIDGE_NAME} -j ACCEPT",
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "ETCD_PORT",
					Value: m.Spec.GiantnetesConfiguration.EtcdPort,
				},
				{
					Name:  "CLUSTER_ID",
					Value: m.Spec.ClusterId,
				},
				{
					Name:  "NETWORK_BRIDGE_NAME",
					Value: networkBridgeName(m.Spec.ClusterId),
				},
				{
					Name: "NODE_IP",
					ValueFrom: &apiv1.EnvVarSource{
						FieldRef: &apiv1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "spec.nodeName",
						},
					},
				},
			},
		},
	}

	bytes, err := json.Marshal(initContainers)
	if err != nil {
		return "", maskAny(err)
	}

	return string(bytes), nil
}

func (m *master) GenerateResources() ([]runtime.Object, error) {
	objects := []runtime.Object{}

	deployment, err := m.GenerateDeployment()
	if err != nil {
		return nil, maskAny(err)
	}

	serviceObjects, err := m.GenerateServiceResources()
	if err != nil {
		return nil, maskAny(err)
	}

	objects = append(objects, deployment)
	objects = append(objects, serviceObjects...)

	return objects, nil
}

func (m *master) GenerateServiceResources() ([]runtime.Object, error) {
	objects := []runtime.Object{}

	endpointMasterEtcd := &extensionsv1.Ingress{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "etcd",
			Labels: map[string]string{
				"cluster":  m.Spec.ClusterId,
				"customer": m.Spec.Customer,
				"app":      "master",
			},
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: "master",
				ServicePort: intstr.FromInt(2379),
			},
		},
	}

	objects = append(objects, endpointMasterEtcd)

	insecurePort, err := strconv.Atoi(m.Spec.Master.InsecurePort)
	if err != nil {
		return nil, maskAny(err)
	}

	endpointMasterAPIHTTP := &extensionsv1.Ingress{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "api",
			Labels: map[string]string{
				"cluster":  m.Spec.ClusterId,
				"customer": m.Spec.Customer,
				"app":      "master",
			},
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: "master",
				ServicePort: intstr.FromInt(insecurePort),
			},
		},
	}

	objects = append(objects, endpointMasterAPIHTTP)
	securePort, err := strconv.Atoi(m.Spec.Master.SecurePort)
	if err != nil {
		return nil, maskAny(err)
	}

	endpointMasterAPIHTTPS := &extensionsv1.Ingress{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "api-https",
			Labels: map[string]string{
				"cluster":  m.Spec.ClusterId,
				"customer": m.Spec.Customer,
				"app":      "master",
			},
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: "master",
				ServicePort: intstr.FromInt(securePort),
			},
		},
	}

	objects = append(objects, endpointMasterAPIHTTPS)

	service := &apiv1.Service{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "master",
			Labels: map[string]string{
				"cluster":  m.Spec.ClusterId,
				"customer": m.Spec.Customer,
				"app":      "master",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceType("LoadBalancer"),
			Ports: []apiv1.ServicePort{
				{
					Name:     "api",
					Port:     int32(insecurePort),
					Protocol: "TCP",
				},
				{
					Name:     "etcd",
					Port:     int32(2379),
					Protocol: "TCP",
				},
				{
					Name:     "api-https",
					Port:     int32(securePort),
					Protocol: "TCP",
				},
			},
		},
	}

	objects = append(objects, service)

	return objects, nil
}

func (m *master) GenerateDeployment() (*extensionsv1.Deployment, error) {
	privileged := true

	initContainers, err := m.generateInitMasterContainers()
	if err != nil {
		return &extensionsv1.Deployment{}, maskAny(err)
	}

	podAffinity, err := m.generateMasterPodAffinity()
	if err != nil {
		return &extensionsv1.Deployment{}, maskAny(err)
	}

	masterReplicas := int32(MasterReplicas)

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "master",
			Labels: map[string]string{
				"cluster":  m.Spec.ClusterId,
				"customer": m.Spec.Customer,
				"app":      "master",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &masterReplicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					GenerateName: "master",
					Labels: map[string]string{
						"cluster":  m.Spec.ClusterId,
						"customer": m.Spec.Customer,
						"app":      "master",
					},
					Annotations: map[string]string{
						"pod.beta.kubernetes.io/init-containers": initContainers,
						"scheduler.alpha.kubernetes.io/affinity": podAffinity,
					},
				},
				Spec: apiv1.PodSpec{
					HostNetwork: true,
					Volumes: []apiv1.Volume{
						{
							Name: "customer-dir",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", m.Spec.ClusterId, "/", m.Spec.ClusterId, "/"),
								},
							},
						},
						{
							Name: "etcd-data",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/home/core/", m.Spec.ClusterId, "-k8s-master-vm/"),
								},
							},
						},
						{
							Name: "api-secrets",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", m.Spec.ClusterId, "/", m.Spec.ClusterId, "/master/secrets"),
								},
							},
						},
						{
							Name: "calico-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", m.Spec.ClusterId, "/", m.Spec.ClusterId, "/ssl/master/calico/"),
								},
							},
						},
						{
							Name: "etcd-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", m.Spec.ClusterId, "/", m.Spec.ClusterId, "/ssl/master/etcd/"),
								},
							},
						},
						{
							Name: "images",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/home/core/images/",
								},
							},
						},
						{
							Name: "rootfs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/home/core/vms/", m.Spec.ClusterId, "-k8s-master-vm/"),
								},
							},
						},
						{
							Name: "ssl",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/ssl/certs/ca-certificates.crt",
								},
							},
						},
						{
							Name: "api-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", m.Spec.ClusterId, "/", m.Spec.ClusterId, "/ssl/master/"),
								},
							},
						},
						{
							Name: "flannel",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/run/flannel",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:            "k8s-vm",
							Image:           fmt.Sprintf("leaseweb-registry.private.giantswarm.io/giantswarm/k8s-vm:%s", m.Spec.K8sVmVersion),
							ImagePullPolicy: apiv1.PullAlways,
							Args: []string{
								"master",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "NETWORK_BRIDGE_NAME",
									Value: networkBridgeName(m.Spec.ClusterId),
								},
								{
									Name:  "CUSTOMER_ID",
									Value: m.Spec.Customer,
								},
								{
									Name:  "G8S_DNS_IP",
									Value: m.Spec.GiantnetesConfiguration.DnsIp,
								},
								{
									Name:  "G8S_DOMAIN",
									Value: m.Spec.GiantnetesConfiguration.Domain,
								},
								{
									Name:  "HOSTNAME",
									Value: m.Spec.ClusterId + "-master.g8s.fra-1.giantswarm.io",
								},
								{
									Name: "HOST_PUBLIC_IP",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
								{
									Name:  "K8S_CLUSTER_IP_RANGE",
									Value: m.Spec.Master.ClusterIpRange,
								},
								{
									Name:  "K8S_CLUSTER_IP_SUBNET",
									Value: m.Spec.Master.ClusterIpSubnet,
								},
								{
									Name: "IP_BRIDGE",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "bridge-ip-configmap-master-vm",
											},
											Key: "bridge-ip",
										},
									},
								},
								{
									Name:  "K8S_INSECURE_PORT",
									Value: m.Spec.Master.InsecurePort,
								},
								{
									Name:  "CALICO_SUBNET",
									Value: m.Spec.Master.CalicoSubnet,
								},
								{
									Name:  "CALICO_CIDR",
									Value: m.Spec.Master.CalicoCidr,
								},
								{
									Name:  "MACHINE_CPU_CORES",
									Value: fmt.Sprintf("%d", m.Spec.Master.Capabilities.CpuCores),
								},
								{
									Name:  "K8S_DNS_IP",
									Value: m.Spec.Master.DnsIp,
								},
								{
									Name:  "K8S_DOMAIN",
									Value: m.Spec.Master.Domain,
								},
								{
									Name:  "K8S_ETCD_DOMAIN_NAME",
									Value: m.Spec.Master.EtcdDomainName,
								},
								{
									Name:  "K8S_ETCD_PREFIX",
									Value: m.Spec.ClusterId,
								},
								{
									Name:  "K8S_MASTER_DOMAIN_NAME",
									Value: m.Spec.Master.MasterDomainName,
								},
								{
									Name:  "K8S_MASTER_SERVICE_NAME",
									Value: m.Spec.Certificates.MasterServiceName,
								},
								{
									Name:  "K8S_NETWORK_SETUP_VERSION",
									Value: m.Spec.Master.NetworkSetupVersion,
								},
								{
									Name:  "DOCKER_EXTRA_ARGS",
									Value: m.Spec.Master.DockerExtraArgs,
								},
								{
									Name:  "K8S_SECURE_PORT",
									Value: m.Spec.Master.SecurePort,
								},
								{
									Name:  "K8S_VERSION",
									Value: m.Spec.K8sVersion,
								},
								{
									Name:  "MACHINE_MEM",
									Value: m.Spec.Master.Capabilities.Memory,
								},
								{
									Name:  "REGISTRY",
									Value: m.Spec.Master.Registry,
								},
								{
									Name: "K8S_ETCD_IP",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "api-certs",
									MountPath: "/etc/kubernetes/ssl/",
								},
								{
									Name:      "api-secrets",
									MountPath: "/etc/kubernetes/secrets/",
								},
								{
									Name:      "images",
									MountPath: "/usr/code/images/",
								},
								{
									Name:      "rootfs",
									MountPath: "/usr/code/rootfs/",
								},
								{
									Name:      "etcd-data",
									MountPath: "/etc/kubernetes/data/etcd/",
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &privileged,
							},
						},
						{
							Name:            "k8s-watch-master-vm",
							Image:           "registry.giantswarm.io/giantswarm/k8s-watch-master-vm:4a226de00c16035f8bb38d43d32b211e5e7d4345",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Env: []apiv1.EnvVar{
								{
									Name:  "CLUSTER_ID",
									Value: m.Spec.ClusterId,
								},
								{
									Name:  "CUSTOMER_ID",
									Value: m.Spec.Customer,
								},
								{
									Name:  "SERVICE_NAME",
									Value: "master",
								},
								{
									Name: "NODE_IP",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
								{
									Name:  "NODE_ETCD_PORT",
									Value: "2379",
								},
								{
									Name:  "G8S_MASTER_HOST",
									Value: "127.0.0.1",
								},
								{
									Name:  "G8S_MASTER_PORT",
									Value: "8080",
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "customer-dir",
									MountPath: "/tmp/",
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &privileged,
							},
						},
					},
				},
			},
		},
	}

	return deployment, nil
}
