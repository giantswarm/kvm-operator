package resources

import (
	"encoding/json"
	"fmt"
	"strconv"

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
	Cluster
}

func (m *master) generateMasterPodAffinity() (string, error) {
	podAffinity := &api.Affinity{
		PodAntiAffinity: &api.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []api.PodAffinityTerm{
				{
					LabelSelector: &apiunversioned.LabelSelector{
						MatchExpressions: []apiunversioned.LabelSelectorRequirement{
							{
								Key:      "role",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{m.Spec.ClusterID + "-worker"},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		},
		PodAffinity: &api.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []api.PodAffinityTerm{
				{
					LabelSelector: &apiunversioned.LabelSelector{
						MatchExpressions: []apiunversioned.LabelSelectorRequirement{
							{
								Key:      "role",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{m.Spec.ClusterID + "-flannel-client"},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
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
			Name:  "set-iptables",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/alpine-bash-iptables",
			Command: []string{
				"/bin/sh",
				"-c",
				"/sbin/iptables -I INPUT -p tcp --match multiport --dports $ETCD_PORT -d ${NODE_IP} -i br${CLUSTER_ID} -j ACCEPT",
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name: "ETCD_PORT",
					ValueFrom: &apiv1.EnvVarSource{
						ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
							LocalObjectReference: apiv1.LocalObjectReference{
								Name: GiantnetesConfigMapName,
							},
							Key: "etcd-port",
						},
					},
				},
				{
					Name:  "CLUSTER_ID",
					Value: m.Spec.ClusterID,
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
		{
			Name:  "generate-bridgeip-configmap",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/generate-bridge-ip-configmap",
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "customer-dir",
					MountPath: "/tmp/",
				},
			},
			Command: []string{
				"/bin/sh",
				"-c",
				"/run.sh",
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "SUFFIX_CONFIGMAP",
					Value: "master-vm",
				},
				{
					Name:  "CUSTOMER_ID",
					Value: m.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: m.Spec.ClusterID,
				},
				{
					Name:  "NAMESPACE",
					Value: m.Name,
				},
			},
		},
		{
			Name:  "kubectl-bridgeip-configmap",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/kubectl:1.4.0",
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "customer-dir",
					MountPath: "/tmp/",
				},
			},
			Command: []string{
				"/bin/sh",
				"-c",
				"while [ ! -f /tmp/bridge-ip-configmap-master-vm.json ]; do echo -; sleep 1; done; /usr/bin/kubectl --server=${G8S_MASTER_HOST}:${G8S_MASTER_PORT} replace --force -f ${BRIDGE_IP_CONFIGMAP_PATH}",
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
					Value: "/tmp/bridge-ip-configmap-master-vm.json",
				},
			},
		},
		{
			Name:  "k8s-master-api-token",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/alpine-openssl",
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
			Name:  "k8s-master-api-certs",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:0.5.0",
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=api.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/master/apiserver.pem --key-file=/etc/kubernetes/ssl/master/apiserver-key.pem --ca-file=/etc/kubernetes/ssl/master/apiserver-ca.pem --alt-names=$K8S_MASTER_SERVICE_NAME,$K8S_API_ALT_NAMES --ip-sans=$G8S_API_IP",
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
					Value: m.Spec.K8sMasterServiceName,
				},
				{
					Name:  "K8S_API_ALT_NAMES",
					Value: m.Spec.K8sAPIaltNames,
				},
				{
					Name: "G8S_API_IP",
					ValueFrom: &apiv1.EnvVarSource{
						ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
							LocalObjectReference: apiv1.LocalObjectReference{
								Name: GiantnetesConfigMapName,
							},
							Key: "g8s-api-ip",
						},
					},
				},
				{
					Name:  "CUSTOMER_ID",
					Value: m.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: m.Spec.ClusterID,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: m.Spec.VaultToken,
				},
				{
					Name: "VAULT_ADDR",
					ValueFrom: &apiv1.EnvVarSource{
						ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
							LocalObjectReference: apiv1.LocalObjectReference{
								Name: GiantnetesConfigMapName,
							},
							Key: "vault-addr",
						},
					},
				},
			},
		},
		{
			Name:  "k8s-master-calico-certs",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:0.5.0",
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=calico.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/calico/client.pem --key-file=/etc/kubernetes/ssl/calico/client-key.pem --ca-file=/etc/kubernetes/ssl/calico/client-ca.pem --alt-names=$K8S_MASTER_SERVICE_NAME",
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
					Value: m.Spec.K8sMasterServiceName,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: m.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: m.Spec.ClusterID,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: m.Spec.VaultToken,
				},
				{
					Name: "VAULT_ADDR",
					ValueFrom: &apiv1.EnvVarSource{
						ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
							LocalObjectReference: apiv1.LocalObjectReference{
								Name: GiantnetesConfigMapName,
							},
							Key: "vault-addr",
						},
					},
				},
			},
		},
		{
			Name:  "k8s-master-etcd-certs",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:0.5.0",
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=etcd.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/etcd/server.pem --key-file=/etc/kubernetes/ssl/etcd/server-key.pem --ca-file=/etc/kubernetes/ssl/etcd/server-ca.pem --alt-names=$K8S_MASTER_SERVICE_NAME",
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
					Value: m.Spec.K8sMasterServiceName,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: m.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: m.Spec.ClusterID,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: m.Spec.VaultToken,
				},
				{
					Name: "VAULT_ADDR",
					ValueFrom: &apiv1.EnvVarSource{
						ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
							LocalObjectReference: apiv1.LocalObjectReference{
								Name: GiantnetesConfigMapName,
							},
							Key: "vault-addr",
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
		return objects, maskAny(err)
	}

	serviceObjects, err := m.GenerateServiceResources()
	if err != nil {
		return objects, maskAny(err)
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
				"cluster-id": m.Spec.ClusterID,
				"role":       m.Spec.ClusterID + "-master",
				"app":        m.Spec.ClusterID + "-k8s-cluster",
			},
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: m.Spec.ClusterID + "-master",
				ServicePort: intstr.FromInt(2379),
			},
		},
	}

	objects = append(objects, endpointMasterEtcd)

	endpointMasterAPIHTTP := &extensionsv1.Ingress{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "api",
			Labels: map[string]string{
				"cluster-id": m.Spec.ClusterID,
				"role":       m.Spec.ClusterID + "-master",
				"app":        m.Spec.ClusterID + "-k8s-cluster",
			},
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: m.Spec.ClusterID + "-master",
				ServicePort: intstr.FromString(m.Spec.K8sInsecurePort),
			},
		},
	}

	objects = append(objects, endpointMasterAPIHTTP)

	endpointMasterAPIHTTPS := &extensionsv1.Ingress{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "api-https",
			Labels: map[string]string{
				"cluster-id": m.Spec.ClusterID,
				"role":       m.Spec.ClusterID + "-master",
				"app":        m.Spec.ClusterID + "-k8s-cluster",
			},
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: m.Spec.ClusterID + "-master",
				ServicePort: intstr.FromString(m.Spec.K8sMasterPort),
			},
		},
	}

	objects = append(objects, endpointMasterAPIHTTPS)

	masterPort, err := strconv.ParseInt(m.Spec.K8sMasterPort, 10, 32)
	if err != nil {
		return []runtime.Object{}, maskAny(err)
	}

	insecurePort, err := strconv.ParseInt(m.Spec.K8sInsecurePort, 10, 32)
	if err != nil {
		return []runtime.Object{}, maskAny(err)
	}

	service := &apiv1.Service{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: m.Spec.ClusterID + "-k8s-master",
			Labels: map[string]string{
				"cluster-id": m.Spec.ClusterID,
				"role":       m.Spec.ClusterID + "-master",
				"app":        m.Spec.ClusterID + "-k8s-cluster",
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
					Port:     int32(masterPort),
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
			Name: m.Spec.ClusterID + "-master",
			Labels: map[string]string{
				"cluster-id": m.Spec.ClusterID,
				"role":       m.Spec.ClusterID + "-master",
				"app":        m.Spec.ClusterID + "-k8s-cluster",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &masterReplicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					GenerateName: m.Spec.ClusterID + "-master",
					Labels: map[string]string{
						"cluster-id": m.Spec.ClusterID,
						"role":       m.Spec.ClusterID + "-master",
						"app":        m.Spec.ClusterID + "-k8s-cluster",
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
									Path: "/etc/kubernetes/" + m.Spec.ClusterID + "/" + m.Spec.ClusterID + "/",
								},
							},
						},
						{
							Name: "etcd-data",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/home/core/" + m.Spec.ClusterID + "-k8s-master-vm/",
								},
							},
						},
						{
							Name: "api-secrets",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + m.Spec.ClusterID + "/" + m.Spec.ClusterID + "/master/secrets",
								},
							},
						},
						{
							Name: "calico-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + m.Spec.ClusterID + "/" + m.Spec.ClusterID + "/ssl/master/calico/",
								},
							},
						},
						{
							Name: "etcd-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + m.Spec.ClusterID + "/" + m.Spec.ClusterID + "/ssl/master/etcd/",
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
									Path: "/home/core/vms/" + m.Spec.ClusterID + "-k8s-master-vm/",
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
									Path: "/etc/kubernetes/" + m.Spec.ClusterID + "/" + m.Spec.ClusterID + "/ssl/master/",
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
							Name:  "k8s-vm",
							Image: fmt.Sprintf("leaseweb-registry.private.giantswarm.io/giantswarm/k8s-vm:%s", m.Spec.K8sVmVersion),
							Args: []string{
								"master",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "BRIDGE_NETWORK",
									Value: "br" + m.Spec.ClusterID,
								},
								{
									Name:  "CUSTOMER_ID",
									Value: m.Spec.Customer,
								},
								{
									Name:  "DOCKER_EXTRA_ARGS",
									Value: m.Spec.DockerExtraArgs,
								},
								{
									Name: "G8S_DNS_IP",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: GiantnetesConfigMapName,
											},
											Key: "g8s-dns-ip",
										},
									},
								},
								{
									Name: "G8S_DOMAIN",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: GiantnetesConfigMapName,
											},
											Key: "g8s-domain",
										},
									},
								},
								{
									Name:  "HOSTNAME",
									Value: m.Spec.ClusterID + "-master.g8s.fra-1.giantswarm.io",
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
									Value: m.Spec.K8sClusterIpRange,
								},
								{
									Name:  "K8S_CLUSTER_IP_SUBNET",
									Value: m.Spec.K8sClusterIpSubnet,
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
									Value: m.Spec.K8sInsecurePort,
								},
								{
									Name:  "K8S_CALICO_MTU",
									Value: m.Spec.K8sCalicoMtu,
								},
								{
									Name:  "CALICO_SUBNET",
									Value: m.Spec.CalicoSubnet,
								},
								{
									Name:  "CALICO_CIDR",
									Value: m.Spec.CalicoCidr,
								},
								{
									Name:  "MACHINE_CPU_CORES",
									Value: fmt.Sprintf("%d", m.Spec.MachineCPUcores),
								},
								{
									Name:  "K8S_DNS_IP",
									Value: m.Spec.K8sDnsIp,
								},
								{
									Name:  "K8S_DOMAIN",
									Value: m.Spec.K8sDomain,
								},
								{
									Name:  "K8S_ETCD_DOMAIN_NAME",
									Value: m.Spec.K8sETCDdomainName,
								},
								{
									Name:  "K8S_ETCD_PREFIX",
									Value: m.Spec.ClusterID,
								},
								{
									Name:  "K8S_MASTER_DOMAIN_NAME",
									Value: m.Spec.K8sMasterDomainName,
								},
								{
									Name:  "K8S_MASTER_PORT",
									Value: m.Spec.K8sMasterPort,
								},
								{
									Name:  "K8S_MASTER_SERVICE_NAME",
									Value: m.Spec.K8sMasterServiceName,
								},
								{
									Name:  "K8S_NETWORK_SETUP_VERSION",
									Value: m.Spec.K8sNetworkSetupVersion,
								},
								{
									Name:  "K8S_NODE_LABELS",
									Value: m.Spec.K8sNodeLabels,
								},
								{
									Name:  "K8S_SECURE_PORT",
									Value: m.Spec.K8sSecurePort,
								},
								{
									Name:  "K8S_VERSION",
									Value: m.Spec.K8sVersion,
								},
								{
									Name:  "MACHINE_MEM",
									Value: m.Spec.MachineMem,
								},
								{
									Name:  "REGISTRY",
									Value: m.Spec.Registry,
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
							Name:  "watch-master-vm-service",
							Image: "leaseweb-registry.private.giantswarm.io/giantswarm/watch-master-vm-service",
							Command: []string{
								"/bin/sh",
								"-c",
								"/run.sh",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "CUSTOMER_ID",
									Value: m.Spec.Customer,
								},
								{
									Name:  "SERVICE_NAME",
									Value: m.Spec.ClusterID + "-k8s-master",
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
