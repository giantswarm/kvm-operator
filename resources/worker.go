package resources

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/giantswarm/kvmtpr"
	"github.com/ventu-io/go-shortid"
	"k8s.io/client-go/pkg/api"
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
)

type Worker interface {
	ClusterObj
}

type worker struct {
	kvmtpr.CustomObject
}

func (w *worker) generateWorkerPodAffinity() (string, error) {
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
					Namespaces:  []string{ClusterID(w.CustomObject)},
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
					Namespaces:  []string{ClusterID(w.CustomObject)},
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

func (w *worker) generateInitWorkerContainers(workerId string) (string, error) {
	privileged := true

	initContainers := []apiv1.Container{
		{
			Name:            "k8s-worker-api-certs",
			Image:           w.Spec.Cluster.Operator.Certctl.Docker.Image,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=$COMMON_NAME --ttl=720h --crt-file=/etc/kubernetes/ssl/" + workerId + "/worker.pem --key-file=/etc/kubernetes/ssl/" + workerId + "/worker-key.pem --ca-file=/etc/kubernetes/ssl/" + workerId + "/worker-ca.pem --alt-names=$ALT_NAMES --ip-sans=$IP_SANS",
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ssl",
					MountPath: "/etc/ssl/certs/ca-certificates.crt",
				},
				{
					Name:      "api-certs",
					MountPath: filepath.Join("/etc/kubernetes/ssl/", workerId, "/"),
				},
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "ALT_NAMES",
					Value: w.Spec.Cluster.Kubernetes.API.AltNames,
				},
				{
					Name:  "CLUSTER_ID",
					Value: ClusterID(w.CustomObject),
				},
				{
					Name:  "COMMON_NAME",
					Value: w.Spec.Cluster.Kubernetes.API.Domain,
				},
				{
					Name:  "IP_SANS",
					Value: w.Spec.Cluster.Kubernetes.API.IP.String(),
				},
				{
					Name:  "VAULT_TOKEN",
					Value: w.Spec.Cluster.Vault.Token,
				},
				{
					Name:  "VAULT_ADDR",
					Value: w.Spec.Cluster.Vault.Address,
				},
			},
		},
		{
			Name:            "k8s-worker-calico-certs",
			Image:           w.Spec.Cluster.Operator.Certctl.Docker.Image,
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
					Value: ClusterID(w.CustomObject),
				},
				{
					Name:  "COMMON_NAME",
					Value: ClusterDomain("calico", ClusterID(w.CustomObject), w.Spec.Cluster.Kubernetes.Domain),
				},
				{
					Name:  "VAULT_TOKEN",
					Value: w.Spec.Cluster.Vault.Token,
				},
				{
					Name:  "VAULT_ADDR",
					Value: w.Spec.Cluster.Vault.Address,
				},
			},
		},
		{
			Name:            "k8s-worker-etcd-certs",
			Image:           w.Spec.Cluster.Operator.Certctl.Docker.Image,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=$COMMON_NAME --ttl=720h --crt-file=/etc/kubernetes/ssl/etcd/client.pem --key-file=/etc/kubernetes/ssl/etcd/client-key.pem --ca-file=/etc/kubernetes/ssl/etcd/client-ca.pem",
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
					Value: ClusterID(w.CustomObject),
				},
				{
					Name:  "COMMON_NAME",
					Value: w.Spec.Cluster.Etcd.Domain,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: w.Spec.Cluster.Vault.Token,
				},
				{
					Name:  "VAULT_ADDR",
					Value: w.Spec.Cluster.Vault.Address,
				},
			},
		},
		{
			Name:            "k8s-endpoint-updater",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-endpoint-updater:84a3506e60edbec199e860070c076948bd9c7ca6",
			ImagePullPolicy: apiv1.PullIfNotPresent,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/k8s-endpoint-updater update --provider.bridge.name=${NETWORK_BRIDGE_NAME} --provider.kind=bridge --service.kubernetes.address=\"\" --service.kubernetes.cluster.namespace=${POD_NAMESPACE} --service.kubernetes.cluster.service=worker --service.kubernetes.inCluster=true --updater.pod.names=${POD_NAME}",
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "NETWORK_BRIDGE_NAME",
					Value: NetworkBridgeName(ClusterID(w.CustomObject)),
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

func (w *worker) GenerateResources() ([]runtime.Object, error) {
	objects := []runtime.Object{}

	workerId, err := generateWorkerId()
	if err != nil {
		return nil, maskAny(err)
	}

	deployment, err := w.GenerateDeployment(workerId)
	if err != nil {
		return nil, maskAny(err)
	}

	service, err := w.GenerateService()
	if err != nil {
		return nil, maskAny(err)
	}

	objects = append(objects, deployment)
	objects = append(objects, service)

	return objects, nil
}

func (w *worker) GenerateService() (*apiv1.Service, error) {
	service := &apiv1.Service{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "worker",
			Labels: map[string]string{
				"cluster":  ClusterID(w.CustomObject),
				"customer": ClusterCustomer(w.CustomObject),
				"app":      "worker",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceTypeLoadBalancer,
			Ports: []apiv1.ServicePort{
				{
					Name:     "http",
					Port:     int32(w.Spec.Cluster.Kubernetes.Kubelet.Port),
					Protocol: "TCP",
				},
			},
			// Note that we do not use a selector definition on purpose to be able to
			// manually set the IP address of the actual VM.
		},
	}

	return service, nil
}

func (w *worker) GenerateDeployment(workerId string) (*extensionsv1.Deployment, error) {
	privileged := true

	initContainers, err := w.generateInitWorkerContainers(workerId)
	if err != nil {
		return nil, maskAny(err)
	}

	podAffinity, err := w.generateWorkerPodAffinity()
	if err != nil {
		return nil, maskAny(err)
	}

	workerReplicas := int32(len(w.Spec.Cluster.Workers))
	workerNode := w.Spec.Cluster.Workers[0]

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "worker",
			Labels: map[string]string{
				"cluster":  ClusterID(w.CustomObject),
				"customer": ClusterCustomer(w.CustomObject),
				"app":      "worker",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: extensionsv1.RecreateDeploymentStrategyType,
			},
			Replicas: &workerReplicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					Name: "worker",
					Labels: map[string]string{
						"cluster":  ClusterID(w.CustomObject),
						"customer": ClusterCustomer(w.CustomObject),
						"app":      "worker",
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
							Name: "api-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", ClusterID(w.CustomObject), "/", ClusterID(w.CustomObject), "/ssl/", workerId, "/"),
								},
							},
						},
						{
							Name: "calico-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", ClusterID(w.CustomObject), "/", ClusterID(w.CustomObject), "/ssl/", workerId, "/calico/"),
								},
							},
						},
						{
							Name: "etcd-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", ClusterID(w.CustomObject), "/", ClusterID(w.CustomObject), "/ssl/", workerId, "/etcd/"),
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
									Path: filepath.Join("/home/core/vms/", ClusterID(w.CustomObject), "-", workerId, "/"),
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
							Name: "certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", ClusterID(w.CustomObject), "/", ClusterID(w.CustomObject), "/ssl/", workerId, "/"),
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "k8s-vm",
							Image: w.Spec.Cluster.Operator.K8sVM.Docker.Image,
							Args: []string{
								"worker",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "CUSTOMER_ID",
									Value: ClusterCustomer(w.CustomObject),
								},
								{
									Name:  "DOCKER_EXTRA_ARGS",
									Value: w.Spec.Cluster.Docker.Daemon.ExtraArgs,
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
									Name:  "K8S_INSECURE_PORT",
									Value: fmt.Sprintf("%d", w.Spec.Cluster.Kubernetes.API.InsecurePort),
								},
								{
									Name:  "K8S_CALICO_MTU",
									Value: w.Spec.Cluster.Calico.MTU,
								},
								{
									Name:  "MACHINE_CPU_CORES",
									Value: fmt.Sprintf("%d", workerNode.CPUs),
								},
								{
									Name:  "NETWORK_BRIDGE_NAME",
									Value: NetworkBridgeName(ClusterID(w.CustomObject)),
								},
								{
									Name:  "K8S_DNS_IP",
									Value: w.Spec.Cluster.Kubernetes.DNS.IP.String(),
								},
								{
									Name:  "K8S_KUBEDNS_DOMAIN",
									Value: ClusterID(w.CustomObject) + ".giantswarm.local.",
								},
								{
									Name:  "K8S_ETCD_DOMAIN_NAME",
									Value: w.Spec.Cluster.Etcd.Domain,
								},
								{
									Name:  "K8S_MASTER_DOMAIN_NAME",
									Value: w.Spec.Cluster.Kubernetes.API.Domain,
								},
								{
									Name:  "K8S_NETWORK_SETUP_IMAGE",
									Value: w.Spec.Cluster.Operator.NetworkSetup.Docker.Image,
								},
								{
									Name:  "K8S_SECURE_PORT",
									Value: fmt.Sprintf("%d", w.Spec.Cluster.Kubernetes.API.SecurePort),
								},
								{
									Name:  "K8S_HYPERKUBE_IMAGE",
									Value: w.Spec.Cluster.Kubernetes.Hyperkube.Docker.Image,
								},
								{
									Name:  "MACHINE_MEM",
									Value: workerNode.Memory,
								},
								{
									Name:  "REGISTRY",
									Value: w.Spec.Cluster.Docker.Registry.Endpoint,
								},
								{
									Name:  "K8S_NODE_LABELS",
									Value: w.Spec.Cluster.Kubernetes.Kubelet.Labels,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "ssl",
									MountPath: "/etc/ssl/certs/ca-certificates.crt",
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
									Name:      "certs",
									MountPath: "/etc/kubernetes/ssl/",
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

func generateWorkerId() (string, error) {
	sid := shortid.GetDefault()

	id, err := sid.Generate()

	return "worker-" + id, err
}
