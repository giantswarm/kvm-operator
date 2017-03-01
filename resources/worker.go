package resources

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/giantswarm/clusterspec"
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
	clusterspec.Cluster
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
					Namespaces:  []string{w.Spec.ClusterId},
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
					Namespaces:  []string{w.Spec.ClusterId},
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
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:" + w.Spec.CertctlVersion,
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
					Value: w.Spec.Certificates.ApiAltNames,
				},
				{
					Name:  "CLUSTER_ID",
					Value: w.Spec.ClusterId,
				},
				{
					Name:  "COMMON_NAME",
					Value: ClusterDomain("api", w.Spec.ClusterId, w.Spec.Worker.Domain),
				},
				{
					Name:  "IP_SANS",
					Value: w.Spec.GiantnetesConfiguration.ApiIp,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: w.Spec.Certificates.VaultToken,
				},
				{
					Name:  "VAULT_ADDR",
					Value: w.Spec.GiantnetesConfiguration.VaultAddr,
				},
			},
		},
		{
			Name:            "k8s-worker-calico-certs",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:" + w.Spec.CertctlVersion,
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
					Value: w.Spec.ClusterId,
				},
				{
					Name:  "COMMON_NAME",
					Value: ClusterDomain("calico", w.Spec.ClusterId, w.Spec.Worker.Domain),
				},
				{
					Name:  "VAULT_TOKEN",
					Value: w.Spec.Certificates.VaultToken,
				},
				{
					Name:  "VAULT_ADDR",
					Value: w.Spec.GiantnetesConfiguration.VaultAddr,
				},
			},
		},
		{
			Name:            "k8s-worker-etcd-certs",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:" + w.Spec.CertctlVersion,
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
					Value: w.Spec.ClusterId,
				},
				{
					Name:  "COMMON_NAME",
					Value: ClusterDomain("etcd", w.Spec.ClusterId, w.Spec.Worker.Domain),
				},
				{
					Name:  "VAULT_TOKEN",
					Value: w.Spec.Certificates.VaultToken,
				},
				{
					Name:  "VAULT_ADDR",
					Value: w.Spec.GiantnetesConfiguration.VaultAddr,
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
	servicePort, err := strconv.ParseInt(w.Spec.Worker.WorkerServicePort, 10, 32)
	if err != nil {
		return nil, maskAny(err)
	}

	service := &apiv1.Service{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "worker",
			Labels: map[string]string{
				"cluster":  w.Spec.ClusterId,
				"customer": w.Spec.Customer,
				"app":      "worker",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceType("NodePort"),
			Ports: []apiv1.ServicePort{
				{
					Name:     "http",
					Port:     int32(servicePort),
					Protocol: "TCP",
				},
			},
			Selector: map[string]string{
				"cluster":  w.Spec.ClusterId,
				"customer": w.Spec.Customer,
				"app":      "worker",
			},
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

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "worker",
			Labels: map[string]string{
				"cluster":  w.Spec.ClusterId,
				"customer": w.Spec.Customer,
				"app":      "worker",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: extensionsv1.RecreateDeploymentStrategyType,
			},
			Replicas: &w.Spec.Worker.Replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					Name: "worker",
					Labels: map[string]string{
						"cluster":  w.Spec.ClusterId,
						"customer": w.Spec.Customer,
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
									Path: filepath.Join("/etc/kubernetes/", w.Spec.ClusterId, "/", w.Spec.ClusterId, "/ssl/", workerId, "/"),
								},
							},
						},
						{
							Name: "calico-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", w.Spec.ClusterId, "/", w.Spec.ClusterId, "/ssl/", workerId, "/calico/"),
								},
							},
						},
						{
							Name: "etcd-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", w.Spec.ClusterId, "/", w.Spec.ClusterId, "/ssl/", workerId, "/etcd/"),
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
									Path: filepath.Join("/home/core/vms/", w.Spec.ClusterId, "-", workerId, "/"),
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
									Path: filepath.Join("/etc/kubernetes/", w.Spec.ClusterId, "/", w.Spec.ClusterId, "/ssl/", workerId, "/"),
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "k8s-vm",
							Image: "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-vm:0f135bdbd732bb78e83abca0bc678e1119ecde99",
							Args: []string{
								"worker",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "CUSTOMER_ID",
									Value: w.Spec.Customer,
								},
								{
									Name:  "DOCKER_EXTRA_ARGS",
									Value: w.Spec.Worker.DockerExtraArgs,
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
									Value: w.Spec.Worker.InsecurePort,
								},
								{
									Name:  "K8S_CALICO_MTU",
									Value: w.Spec.Worker.K8sCalicoMtu,
								},
								{
									Name:  "MACHINE_CPU_CORES",
									Value: fmt.Sprintf("%d", w.Spec.Worker.Capabilities.CpuCores),
								},
								{
									Name:  "NETWORK_BRIDGE_NAME",
									Value: NetworkBridgeName(w.Spec.ClusterId),
								},
								{
									Name:  "K8S_DNS_IP",
									Value: w.Spec.Worker.DnsIp,
								},
								{
									Name:  "K8S_DOMAIN", // TODO rename to K8S_KUBEDNS_DOMAIN
									Value: w.Spec.ClusterId + ".giantswarm.local.",
								},
								{
									Name:  "K8S_ETCD_DOMAIN_NAME",
									Value: w.Spec.Worker.EtcdDomainName,
								},
								{
									Name:  "K8S_MASTER_DOMAIN_NAME",
									Value: w.Spec.Worker.MasterDomainName,
								},
								{
									Name:  "K8S_NETWORK_SETUP_VERSION",
									Value: w.Spec.Worker.NetworkSetupVersion,
								},
								{
									Name:  "K8S_SECURE_PORT",
									Value: w.Spec.Worker.SecurePort,
								},
								{
									Name:  "K8S_VERSION",
									Value: w.Spec.K8sVersion,
								},
								{
									Name:  "MACHINE_MEM",
									Value: w.Spec.Worker.Capabilities.Memory,
								},
								{
									Name:  "REGISTRY",
									Value: w.Spec.Worker.Registry,
								},
								{
									Name:  "K8S_NODE_LABELS",
									Value: w.Spec.Worker.NodeLabels,
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
