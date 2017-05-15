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
			Name:            "set-iptables",
			Image:           m.Spec.KVM.Network.IPTables.Docker.Image,
			ImagePullPolicy: apiv1.PullIfNotPresent,
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
		{
			Name:            "k8s-endpoint-updater",
			Image:           m.Spec.KVM.EndpointUpdater.Docker.Image,
			ImagePullPolicy: apiv1.PullIfNotPresent,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/k8s-endpoint-updater update --provider.bridge.name=${NETWORK_BRIDGE_NAME} --provider.kind=bridge --service.kubernetes.address=\"\" --service.kubernetes.cluster.namespace=${POD_NAMESPACE} --service.kubernetes.cluster.service=master --service.kubernetes.inCluster=true --updater.pod.names=${POD_NAME}",
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
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
			Type: apiv1.ServiceTypeLoadBalancer,
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
			// Note that we do not use a selector definition on purpose to be able to
			// manually set the IP address of the actual VM.
		},
	}

	objects = append(objects, service)

	return objects, nil
}

func (m *master) GenerateDeployment() (*extensionsv1.Deployment, error) {
	privileged := true

	initContainers, err := m.generateInitMasterContainers()
	if err != nil {
		return nil, maskAny(err)
	}

	podAffinity, err := m.generateMasterPodAffinity()
	if err != nil {
		return nil, maskAny(err)
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
							Name:            "k8s-kvm",
							Image:           m.Spec.KVM.K8sKVM.Docker.Image,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								"master",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "CORES",
									Value: fmt.Sprintf("%d", masterNode.CPUs),
								},
								{
									Name:  "DISK",
									Value: "4G",
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
									Name:  "NETWORK_BRIDGE_NAME",
									Value: NetworkBridgeName(ClusterID(m.CustomObject)),
								},
								{
									Name:  "MEMORY",
									Value: masterNode.Memory,
								},
								{
									Name:  "ROLE",
									Value: "master",
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
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
								// TODO cloud config has to be written into "/usr/code/cloudconfig/openstack/latest/user_data".
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
