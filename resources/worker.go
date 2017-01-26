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
					Value: bridgeIPConfigmapName("worker"),
				},
				{
					Name:  "BRIDGE_IP_CONFIGMAP_PATH",
					Value: bridgeIPConfigmapPath("worker"),
				},
				{
					Name:  "K8S_NAMESPACE",
					Value: w.Spec.ClusterId,
				},
				{
					Name:  "NETWORK_BRIDGE_NAME",
					Value: networkBridgeName(w.Spec.ClusterId),
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
					Value: bridgeIPConfigmapPath("worker"),
				},
			},
		},
		{
			Name:            "k8s-worker-api-certs",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:" + w.Spec.CertctlVersion,
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=api.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/" + workerId + "/worker.pem --key-file=/etc/kubernetes/ssl/" + workerId + "/worker-key.pem --ca-file=/etc/kubernetes/ssl/" + workerId + "/worker-ca.pem --alt-names=$K8S_MASTER_SERVICE_NAME,$K8S_API_ALT_NAMES --ip-sans=$G8S_API_IP",
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
					Name:  "K8S_MASTER_SERVICE_NAME",
					Value: w.Spec.Certificates.MasterServiceName,
				},
				{
					Name:  "K8S_API_ALT_NAMES",
					Value: w.Spec.Certificates.ApiAltNames,
				},
				{
					Name:  "G8S_API_IP",
					Value: w.Spec.GiantnetesConfiguration.ApiIp,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: w.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: w.Spec.ClusterId,
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
					Value: w.Spec.Certificates.MasterServiceName,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: w.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: w.Spec.ClusterId,
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
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=etcd.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/etcd/client.pem --key-file=/etc/kubernetes/ssl/etcd/client-key.pem --ca-file=/etc/kubernetes/ssl/etcd/client-ca.pem --alt-names=$K8S_MASTER_SERVICE_NAME",
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
					Value: w.Spec.Certificates.MasterServiceName,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: w.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: w.Spec.ClusterId,
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
				Type: "Recreate",
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
							Name: "customer-dir",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", w.Spec.ClusterId, "/", w.Spec.ClusterId, "/"),
								},
							},
						},
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
							Name: "bridge-ip-configmap",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: filepath.Join("/etc/kubernetes/", w.Spec.ClusterId, "/", w.Spec.ClusterId, "/"),
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
							Name:  "vm",
							Image: "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-vm:0868cdd0b0c7bf3b01fc108d7b50436bbdc4a65e",
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
									Name:  "G8S_DNS_IP",
									Value: w.Spec.GiantnetesConfiguration.DnsIp,
								},
								{
									Name:  "G8S_DOMAIN",
									Value: w.Spec.GiantnetesConfiguration.Domain,
								},
								{
									Name:  "HOSTNAME",
									Value: w.Spec.ClusterId + "-k8svm-" + workerId,
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
									Name: "IP_BRIDGE",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: bridgeIPConfigmapName("worker"),
											},
											Key: "bridge-ip",
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
									Value: networkBridgeName(w.Spec.ClusterId),
								},
								{
									Name:  "K8S_DNS_IP",
									Value: w.Spec.Worker.DnsIp,
								},
								{
									Name:  "K8S_DOMAIN",
									Value: w.Spec.Worker.Domain,
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
									Name:  "K8S_MASTER_PORT",
									Value: w.Spec.Worker.MasterPort,
								},
								{
									Name:  "K8S_MASTER_SERVICE_NAME",
									Value: w.Spec.Certificates.MasterServiceName,
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
