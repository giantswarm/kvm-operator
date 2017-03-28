package resources

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/giantswarm/kvmtpr"
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
	kvmtpr.CustomObject
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
					Namespaces:  []string{ClusterID(m.CustomObject)},
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
					Namespaces:  []string{ClusterID(m.CustomObject)},
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
			Name:            "k8s-master-api-token",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-network-openssl:410c14100b89ffad9d84f0a5fbd9bdb398cdc2fd",
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
			Image:           m.Spec.Cluster.Operator.Certctl.Docker.Image,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=$COMMON_NAME --ttl=720h --crt-file=/etc/kubernetes/ssl/master/apiserver.pem --key-file=/etc/kubernetes/ssl/master/apiserver-key.pem --ca-file=/etc/kubernetes/ssl/master/apiserver-ca.pem --alt-names=$ALT_NAMES --ip-sans=$IP_SANS",
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
					Name:  "ALT_NAMES",
					Value: m.Spec.Cluster.Kubernetes.API.AltNames,
				},
				{
					Name:  "COMMON_NAME",
					Value: m.Spec.Cluster.Kubernetes.API.Domain,
				},
				{
					Name:  "CLUSTER_ID",
					Value: ClusterID(m.CustomObject),
				},
				{
					Name:  "IP_SANS",
					Value: m.Spec.Cluster.Kubernetes.API.IP.String(),
				},
				{
					Name:  "VAULT_TOKEN",
					Value: m.Spec.Cluster.Vault.Token,
				},
				{
					Name:  "VAULT_ADDR",
					Value: m.Spec.Cluster.Vault.Address,
				},
			},
		},
		{
			Name:            "k8s-master-calico-certs",
			Image:           m.Spec.Cluster.Operator.Certctl.Docker.Image,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=$COMMON_NAME --ttl=720h --crt-file=/etc/kubernetes/ssl/calico/client.pem --key-file=/etc/kubernetes/ssl/calico/client-key.pem --ca-file=/etc/kubernetes/ssl/calico/client-ca.pem",
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
					Name:  "CLUSTER_ID",
					Value: ClusterID(m.CustomObject),
				},
				{
					Name:  "COMMON_NAME",
					Value: ClusterDomain("calico", ClusterID(m.CustomObject), m.Spec.Cluster.Kubernetes.Domain),
				},
				{
					Name:  "VAULT_TOKEN",
					Value: m.Spec.Cluster.Vault.Token,
				},
				{
					Name:  "VAULT_ADDR",
					Value: m.Spec.Cluster.Vault.Address,
				},
			},
		},
		{
			Name:            "k8s-master-etcd-certs",
			Image:           m.Spec.Cluster.Operator.Certctl.Docker.Image,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=$COMMON_NAME --ttl=720h --crt-file=/etc/kubernetes/ssl/etcd/server.pem --key-file=/etc/kubernetes/ssl/etcd/server-key.pem --ca-file=/etc/kubernetes/ssl/etcd/server-ca.pem",
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
					Name:  "CLUSTER_ID",
					Value: ClusterID(m.CustomObject),
				},
				{
					Name:  "COMMON_NAME",
					Value: m.Spec.Cluster.Etcd.Domain,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: m.Spec.Cluster.Vault.Token,
				},
				{
					Name:  "VAULT_ADDR",
					Value: m.Spec.Cluster.Vault.Address,
				},
			},
		},
		{
			Name:            "set-iptables",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-network-iptables:4625e26b128c0ce637774ab0a3051fb6df07d0be",
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/sbin/iptables -I INPUT -p tcp --match multiport --dports ${ETCD_PORT} -d ${NODE_IP} -i ${NETWORK_BRIDGE_NAME} -j ACCEPT",
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "ETCD_PORT",
					Value: fmt.Sprintf("%d", m.Spec.Cluster.Etcd.Port),
				},
				{
					Name:  "NETWORK_BRIDGE_NAME",
					Value: NetworkBridgeName(ClusterID(m.CustomObject)),
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
				"cluster":  ClusterID(m.CustomObject),
				"customer": ClusterCustomer(m.CustomObject),
				"app":      "master",
			},
			Annotations: map[string]string{
				"ingress.kubernetes.io/ssl-passthrough": "true",
			},
		},
		Spec: extensionsv1.IngressSpec{
			TLS: []extensionsv1.IngressTLS{
				extensionsv1.IngressTLS{
					Hosts: []string{
						m.Spec.Cluster.Etcd.Domain,
					},
				},
			},
			Rules: []extensionsv1.IngressRule{
				extensionsv1.IngressRule{
					Host: m.Spec.Cluster.Etcd.Domain,
					IngressRuleValue: extensionsv1.IngressRuleValue{
						HTTP: &extensionsv1.HTTPIngressRuleValue{
							Paths: []extensionsv1.HTTPIngressPath{
								extensionsv1.HTTPIngressPath{
									Path: "/",
									Backend: extensionsv1.IngressBackend{
										ServiceName: "master",
										ServicePort: intstr.FromInt(2379),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	objects = append(objects, endpointMasterEtcd)

	endpointMasterAPIHTTPS := &extensionsv1.Ingress{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "api",
			Labels: map[string]string{
				"cluster":  ClusterID(m.CustomObject),
				"customer": ClusterCustomer(m.CustomObject),
				"app":      "master",
			},
			Annotations: map[string]string{
				"ingress.kubernetes.io/ssl-passthrough": "true",
			},
		},
		Spec: extensionsv1.IngressSpec{
			TLS: []extensionsv1.IngressTLS{
				extensionsv1.IngressTLS{
					Hosts: []string{
						m.Spec.Cluster.Kubernetes.API.Domain,
					},
				},
			},
			Rules: []extensionsv1.IngressRule{
				extensionsv1.IngressRule{
					Host: m.Spec.Cluster.Kubernetes.API.Domain,
					IngressRuleValue: extensionsv1.IngressRuleValue{
						HTTP: &extensionsv1.HTTPIngressRuleValue{
							Paths: []extensionsv1.HTTPIngressPath{
								extensionsv1.HTTPIngressPath{
									Path: "/",
									Backend: extensionsv1.IngressBackend{
										ServiceName: "master",
										ServicePort: intstr.FromInt(m.Spec.Cluster.Kubernetes.API.SecurePort),
									},
								},
							},
						},
					},
				},
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
				"cluster":  ClusterID(m.CustomObject),
				"customer": ClusterCustomer(m.CustomObject),
				"app":      "master",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceType("LoadBalancer"),
			Ports: []apiv1.ServicePort{
				{
					Name:     "etcd",
					Port:     int32(2379),
					Protocol: "TCP",
				},
				{
					Name:     "api",
					Port:     int32(m.Spec.Cluster.Kubernetes.API.SecurePort),
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
	masterNode := m.Spec.Cluster.Masters[0]

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "master",
			Labels: map[string]string{
				"cluster":  ClusterID(m.CustomObject),
				"customer": ClusterCustomer(m.CustomObject),
				"app":      "master",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: extensionsv1.RecreateDeploymentStrategyType,
			},
			Replicas: &masterReplicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					GenerateName: "master",
					Labels: map[string]string{
						"cluster":  ClusterID(m.CustomObject),
						"customer": ClusterCustomer(m.CustomObject),
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
							Name: "etcd-data",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/home/core/", ClusterID(m.CustomObject), "-k8s-master-vm/"),
								},
							},
						},
						{
							Name: "api-secrets",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", ClusterID(m.CustomObject), "/", ClusterID(m.CustomObject), "/master/secrets"),
								},
							},
						},
						{
							Name: "calico-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", ClusterID(m.CustomObject), "/", ClusterID(m.CustomObject), "/ssl/master/calico/"),
								},
							},
						},
						{
							Name: "etcd-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", ClusterID(m.CustomObject), "/", ClusterID(m.CustomObject), "/ssl/master/etcd/"),
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
									Path: filepath.Join("/home/core/vms/", ClusterID(m.CustomObject), "-k8s-master-vm/"),
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
									Path: filepath.Join("/etc/kubernetes/", ClusterID(m.CustomObject), "/", ClusterID(m.CustomObject), "/ssl/master/"),
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
							Image:           m.Spec.Cluster.Operator.K8sVM.Docker.Image,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								"master",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "NETWORK_BRIDGE_NAME",
									Value: NetworkBridgeName(ClusterID(m.CustomObject)),
								},
								{
									Name:  "CUSTOMER_ID",
									Value: ClusterCustomer(m.CustomObject),
								},
								{
									Name: "HOSTNAME",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
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
									Value: m.Spec.Cluster.Kubernetes.API.ClusterIPRange,
								},
								{
									Name:  "K8S_CLUSTER_IP_SUBNET",
									Value: m.Spec.Cluster.Kubernetes.API.ClusterIPRange,
								},
								{
									Name:  "K8S_INSECURE_PORT",
									Value: fmt.Sprintf("%d", m.Spec.Cluster.Kubernetes.API.InsecurePort),
								},
								{
									Name:  "CALICO_SUBNET",
									Value: m.Spec.Cluster.Calico.Subnet,
								},
								{
									Name:  "CALICO_CIDR",
									Value: m.Spec.Cluster.Calico.CIDR,
								},
								{
									Name:  "MACHINE_CPU_CORES",
									Value: fmt.Sprintf("%d", masterNode.CPUs),
								},
								{
									Name:  "K8S_DNS_IP",
									Value: m.Spec.Cluster.Kubernetes.DNS.IP.String(),
								},
								{
									Name:  "K8S_KUBEDNS_DOMAIN",
									Value: ClusterID(m.CustomObject) + ".giantswarm.local.",
								},
								{
									Name:  "K8S_ETCD_DOMAIN_NAME",
									Value: m.Spec.Cluster.Etcd.Domain,
								},
								{
									Name:  "K8S_ETCD_PREFIX",
									Value: ClusterID(m.CustomObject),
								},
								{
									Name:  "K8S_MASTER_DOMAIN_NAME",
									Value: m.Spec.Cluster.Kubernetes.API.Domain,
								},
								{
									Name:  "K8S_NETWORK_SETUP_IMAGE",
									Value: m.Spec.Cluster.Operator.NetworkSetup.Docker.Image,
								},
								{
									Name:  "DOCKER_EXTRA_ARGS",
									Value: m.Spec.Cluster.Docker.Daemon.ExtraArgs,
								},
								{
									Name:  "K8S_SECURE_PORT",
									Value: fmt.Sprintf("%d", m.Spec.Cluster.Kubernetes.API.SecurePort),
								},
								{
									Name:  "K8S_HYPERKUBE_IMAGE",
									Value: m.Spec.Cluster.Kubernetes.Hyperkube.Docker.Image,
								},
								{
									Name:  "MACHINE_MEM",
									Value: masterNode.Memory,
								},
								{
									Name:  "REGISTRY",
									Value: m.Spec.Cluster.Docker.Registry.Endpoint,
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
							Name:            "k8s-endpoint-updater",
							Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-endpoint-updater:a68e1c976c2cae116e7e12429dcd2b8f5f226c45",
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Command: []string{
								"/bin/sh",
								"-c",
								"/opt/k8s-endpoint-updater update --provider.bridge.name=${NETWORK_BRIDGE_NAME} --provider.kind=bridge --service.kubernetes.cluster.namespace=${POD_NAMESPACE} --service.kubernetes.inCluster=true --updater.pod.names=${POD_NAME}",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "NETWORK_BRIDGE_NAME",
									Value: NetworkBridgeName(ClusterID(m.CustomObject)),
								},
								{
									Name: "POD_NAME",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
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
