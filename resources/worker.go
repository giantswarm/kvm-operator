package resources

import (
	"encoding/json"
	"fmt"
	"strconv"

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
	Cluster
}

func (w *worker) generateWorkerPodAffinity() (string, error) {
	podAffinity := &api.Affinity{
		PodAntiAffinity: &api.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []api.PodAffinityTerm{
				{
					LabelSelector: &apiunversioned.LabelSelector{
						MatchExpressions: []apiunversioned.LabelSelectorRequirement{
							{
								Key:      "role",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{w.Name + "-master"},
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
								Values:   []string{w.Name + "-flannel-client"},
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

func (w *worker) generateInitWorkerContainers(workerId string) (string, error) {
	privileged := true

	initContainers := []apiv1.Container{
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
					Value: "worker-vm",
				},
				{
					Name:  "CUSTOMER_ID",
					Value: w.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: w.Spec.ClusterID,
				},
				{
					Name:  "NAMESPACE",
					Value: w.Name,
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
				"while [ ! -f /tmp/bridge-ip-configmap-worker-vm.json ]; do echo -; sleep 1; done; /usr/bin/kubectl --server=${G8S_MASTER_HOST}:${G8S_MASTER_PORT} replace --force -f ${BRIDGE_IP_CONFIGMAP_PATH}",
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
					Value: "/tmp/bridge-ip-configmap-worker-vm.json",
				},
			},
		},
		{
			Name:  "k8s-worker-api-certs",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:0.5.0",
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
					MountPath: "/etc/kubernetes/ssl/" + workerId + "/",
				},
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "K8S_MASTER_SERVICE_NAME",
					Value: w.Spec.K8sMasterServiceName,
				},
				{
					Name:  "K8S_API_ALT_NAMES",
					Value: w.Spec.K8sAPIaltNames,
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
					Value: w.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: w.Spec.ClusterID,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: w.Spec.VaultToken,
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
			Name:  "k8s-worker-calico-certs",
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
					Value: w.Spec.K8sMasterServiceName,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: w.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: w.Spec.ClusterID,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: w.Spec.VaultToken,
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
			Name:  "k8s-worker-etcd-certs",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/certctl:0.5.0",
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
					Value: w.Spec.K8sMasterServiceName,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: w.Spec.Customer,
				},
				{
					Name:  "CLUSTER_ID",
					Value: w.Spec.ClusterID,
				},
				{
					Name:  "VAULT_TOKEN",
					Value: w.Spec.VaultToken,
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

func (w *worker) GenerateResources() ([]runtime.Object, error) {
	objects := []runtime.Object{}

	workerId, err := generateWorkerId()
	if err != nil {
		return objects, maskAny(err)
	}

	deployment, err := w.GenerateDeployment(workerId)
	if err != nil {
		return objects, maskAny(err)
	}

	service, err := w.GenerateService()
	if err != nil {
		return objects, maskAny(err)
	}

	objects = append(objects, deployment)
	objects = append(objects, service)

	return objects, nil
}

func (w *worker) GenerateService() (*apiv1.Service, error) {
	servicePort, err := strconv.ParseInt(w.Spec.K8sWorkerServicePort, 10, 32)
	if err != nil {
		return nil, maskAny(err)
	}

	service := &apiv1.Service{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: w.Spec.ClusterID + "-worker",
			Labels: map[string]string{
				"cluster-id": w.Spec.ClusterID,
				"role":       w.Spec.ClusterID + "-worker",
				"app":        w.Spec.ClusterID + "-k8s-cluster",
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
				"app":  w.Spec.ClusterID + "-k8s-cluster",
				"role": "worker",
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
			Name: w.Spec.ClusterID + "-worker",
			Labels: map[string]string{
				"cluster-id": w.Spec.ClusterID,
				"role":       w.Spec.ClusterID + "-worker",
				"app":        "k8s-cluster",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &w.Spec.WorkerReplicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					Name: w.Spec.ClusterID + "-worker",
					Labels: map[string]string{
						"cluster-id": w.Spec.ClusterID,
						"role":       w.Spec.ClusterID + "-worker",
						"app":        "k8s-cluster",
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
									Path: "/etc/kubernetes/" + w.Spec.ClusterID + "/" + w.Spec.ClusterID + "/",
								},
							},
						},
						{
							Name: "api-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + w.Spec.ClusterID + "/" + w.Spec.ClusterID + "/ssl/" + workerId + "/",
								},
							},
						},
						{
							Name: "calico-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + w.Spec.ClusterID + "/" + w.Spec.ClusterID + "/ssl/" + workerId + "/calico/",
								},
							},
						},
						{
							Name: "etcd-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + w.Spec.ClusterID + "/" + w.Spec.ClusterID + "/ssl/" + workerId + "/etcd/",
								},
							},
						},
						{
							Name: "bridge-ip-configmap",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + w.Spec.ClusterID + "/" + w.Spec.ClusterID + "/",
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
									Path: "/home/core/vms/" + w.Spec.ClusterID + "-" + workerId + "/",
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
									Path: "/etc/kubernetes/" + w.Spec.ClusterID + "/" + w.Spec.ClusterID + "/ssl/" + workerId + "/",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "vm",
							Image: "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-vm:0.9.13",
							Args: []string{
								"worker",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "BRIDGE_NETWORK",
									Value: "br" + w.Spec.ClusterID,
								},
								{
									Name:  "CUSTOMER_ID",
									Value: w.Spec.Customer,
								},
								{
									Name:  "DOCKER_EXTRA_ARGS",
									Value: w.Spec.DockerExtraArgs,
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
									Value: w.Spec.ClusterID + "-k8svm-" + workerId,
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
									Name:  "IP_BRIDGE",
									Value: "",
								},
								{
									Name:  "K8S_INSECURE_PORT",
									Value: w.Spec.K8sInsecurePort,
								},
								{
									Name:  "K8S_CALICO_MTU",
									Value: w.Spec.K8sCalicoMtu,
								},
								{
									Name:  "MACHINE_CPU_CORES",
									Value: fmt.Sprintf("%d", w.Spec.MachineCPUcores),
								},
								{
									Name:  "K8S_DNS_IP",
									Value: w.Spec.K8sDnsIp,
								},
								{
									Name:  "K8S_DOMAIN",
									Value: w.Spec.K8sDomain,
								},
								{
									Name:  "K8S_ETCD_DOMAIN_NAME",
									Value: w.Spec.K8sETCDdomainName,
								},
								{
									Name:  "K8S_ETCD_PREFIX",
									Value: w.Spec.ClusterID,
								},
								{
									Name:  "K8S_MASTER_DOMAIN_NAME",
									Value: w.Spec.K8sMasterDomainName,
								},
								{
									Name:  "K8S_MASTER_PORT",
									Value: w.Spec.K8sMasterPort,
								},
								{
									Name:  "K8S_MASTER_SERVICE_NAME",
									Value: w.Spec.K8sMasterServiceName,
								},
								{
									Name:  "K8S_NETWORK_SETUP_VERSION",
									Value: w.Spec.K8sNetworkSetupVersion,
								},
								{
									Name:  "K8S_NODE_LABELS",
									Value: w.Spec.K8sNodeLabels,
								},
								{
									Name:  "K8S_SECURE_PORT",
									Value: w.Spec.K8sSecurePort,
								},
								{
									Name:  "K8S_VERSION",
									Value: w.Spec.K8sVersion,
								},
								{
									Name:  "MACHINE_MEM",
									Value: w.Spec.MachineMem,
								},
								{
									Name:  "REGISTRY",
									Value: w.Spec.Registry,
								},
								{
									Name:  "DOCKER_EXTRA_ARGS",
									Value: w.Spec.DockerExtraArgs,
								},
								{
									Name:  "K8S_NODE_LABELS",
									Value: w.Spec.K8sNodeLabels,
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
