package resources

import (
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"

	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
)

const podAffinityWorker string = `
{
	"podAffinity": {
		"requiredDuringSchedulingIgnoredDuringExecution": [
			{
				"labelSelector": {
					"matchExpressions": [
						{
							"key": "role",
							"operator": "In",
							"values": ["{{.ClusterID}}-flannel-client"]
						}
					]
				},
				"topologyKey": "kubernetes.io/hostname"
		 }
		]
	 },
	"podAntiAffinity": {
		"requiredDuringSchedulingIgnoredDuringExecution": [
			{
				"labelSelector": {
					"matchExpressions": [
						{
							"key": "role",
							"operator": "In",
							"values": ["{{.ClusterID}}-master"]
						}
					]
				},
				"topologyKey": "kubernetes.io/hostname"
		 }
		]
	 }
 }`

const initContainersWorker string = `[
			{
				"name": "generate-bridgeip-configmap",
				"image": "hectorj2f/generate-bridge-ip-configmap",
				"securityContext": {
					"privileged": true
				},
				"imagePullPolicy": "Always",
				"volumeMounts": [
					{
						"name": "bridge-ip-configmap",
						"mountPath": "/tmp/"
					}
				],
				"env": [
					{
						"name": "CLUSTER_ID",
						"valueFrom": {
							"configMapKeyRef": {
								"name": "{{.ClusterID}}-configmap",
								"key": "cluster-id"
							}
						}
					},
					{
						"name": "CUSTOMER_ID",
						"valueFrom": {
							"configMapKeyRef": {
								"name": "{{.ClusterID}}-configmap",
								"key": "customer-id"
							}
						}
					},
					{
						"name": "SUFFIX_CONFIGMAP",
						"value": "worker-vm"
					}
				],
				"command": [
					"/bin/sh",
					"-c",
					"/run.sh"
				]
			},
			{
				"name": "kubectl-bridgeip-configmap",
				"image": "hectorj2f/kubectl:1.4.0",
				"imagePullPolicy": "Always",
				"volumeMounts": [
					{
						"name": "bridge-ip-configmap",
						"mountPath": "/tmp/"
					}
				],
				"env": [
					{
						"name": "G8S_MASTER_PORT",
						"value": "8080"
					},
					{
						"name": "G8S_MASTER_HOST",
						"value": "127.0.0.1"
					},
					{
						"name": "BRIDGE_IP_CONFIGMAP_PATH",
						"value": "/tmp/bridge-ip-configmap-worker-vm.json"
					}
				],
				"command": [
					"/bin/sh",
					"-c",
					"while [ ! -f /tmp/bridge-ip-configmap-worker-vm.json ]; do echo _; sleep 1; done; /usr/bin/kubectl --server=${G8S_MASTER_HOST}:${G8S_MASTER_PORT} replace --force -f ${BRIDGE_IP_CONFIGMAP_PATH}"
				]
		},
		{
			"name": "k8s-worker-api-certs",
			"securityContext": {
				"privileged": true
			},
			"image": "giantswarm/certctl:0.5.0",
			"imagePullPolicy": "IfNotPresent",
			"volumeMounts": [
				{
					"name": "api-certs",
					"mountPath": "/etc/kubernetes/ssl/"
				},
				{
					"name": "ssl",
					"mountPath": "/etc/ssl/certs/ca-certificates.crt"
				}
			],
			"command": [
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=api.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/worker.pem --key-file=/etc/kubernetes/ssl/worker-key.pem --ca-file=/etc/kubernetes/ssl/worker-ca.pem --alt-names=$K8S_MASTER_SERVICE_NAME,$K8S_API_ALT_NAMES --ip-sans=$G8S_API_IP"
			],
			"env": [
				{
					"name": "VAULT_ADDR",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "vault-addr"
						}
					}
				},
				{
					"name": "VAULT_TOKEN",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "vault-token"
						}
					}
				},
				{
					"name": "CLUSTER_ID",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "cluster-id"
						}
					}
				},
				{
					"name": "CUSTOMER_ID",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "customer-id"
						}
					}
				},
				{
					"name": "G8S_API_IP",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "g8s-api-ip"
						}
					}
				},
				{
					"name": "K8S_API_ALT_NAMES",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "k8s-api-alt-names"
						}
					}
				},
				{
					"name": "K8S_MASTER_SERVICE_NAME",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "k8s-master-service-name"
						}
					}
				}
			]
		},
		{
			"name": "k8s-worker-etcd-certs",
			"securityContext": {
				"privileged": true
			},
			"image": "giantswarm/certctl:0.5.0",
			"imagePullPolicy": "IfNotPresent",
			"volumeMounts": [
				{
					"name": "etcd-certs",
					"mountPath": "/etc/kubernetes/ssl/etcd/"
				},
				{
					"name": "ssl",
					"mountPath": "/etc/ssl/certs/ca-certificates.crt"
				}
			],
			"command": [
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=etcd.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/etcd/client.pem --key-file=/etc/kubernetes/ssl/etcd/client-key.pem --ca-file=/etc/kubernetes/ssl/etcd/client-ca.pem --alt-names=$K8S_MASTER_SERVICE_NAME"
			],
			"env": [
				{
					"name": "VAULT_ADDR",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "vault-addr"
						}
					}
				},
				{
					"name": "VAULT_TOKEN",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "vault-token"
						}
					}
				},
				{
					"name": "CLUSTER_ID",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "cluster-id"
						}
					}
				},
				{
					"name": "CUSTOMER_ID",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "customer-id"
						}
					}
				},
				{
					"name": "K8S_MASTER_SERVICE_NAME",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "k8s-master-service-name"
						}
					}
				}
			]
		},
		{
			"name": "k8s-worker-calico-certs",
			"securityContext": {
				"privileged": true
			},
			"image": "giantswarm/certctl:0.5.0",
			"imagePullPolicy": "IfNotPresent",
			"volumeMounts": [
				{
					"name": "calico-certs",
					"mountPath": "/etc/kubernetes/ssl/calico/"
				},
				{
					"name": "ssl",
					"mountPath": "/etc/ssl/certs/ca-certificates.crt"
				}
			],
			"command": [
				"/bin/sh",
				"-c",
				"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=calico.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/calico/client.pem --key-file=/etc/kubernetes/ssl/calico/client-key.pem --ca-file=/etc/kubernetes/ssl/calico/client-ca.pem --alt-names=$K8S_MASTER_SERVICE_NAME"
			],
			"env": [
				{
					"name": "VAULT_ADDR",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "vault-addr"
						}
					}
				},
				{
					"name": "VAULT_TOKEN",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "vault-token"
						}
					}
				},
				{
					"name": "CLUSTER_ID",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "cluster-id"
						}
					}
				},
				{
					"name": "CUSTOMER_ID",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "customer-id"
						}
					}
				},
				{
					"name": "K8S_MASTER_SERVICE_NAME",
					"valueFrom": {
						"configMapKeyRef": {
							"name": "{{.ClusterID}}-configmap",
							"key": "k8s-master-service-name"
						}
					}
				}
			]
		}
]`

type Worker interface {
	ClusterObj
}

type worker struct {
	Cluster
}

func (w *worker) GenerateResources() ([]runtime.Object, error) {
	objects := []runtime.Object{}

	deployment, err := w.GenerateDeployment()
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
	service := &apiv1.Service{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apiv1.ObjectMeta{
			GenerateName: w.Spec.ClusterID + "-worker-vm",
			Labels: map[string]string{
				"cluster-id": w.Spec.ClusterID,
				"role":       w.Spec.ClusterID + "worker",
				"app":        w.Spec.ClusterID + "-k8s-cluster",
			},
		},
		Spec: apiv1.ServiceSpec{
			Type: apiv1.ServiceType("NodePort"),
			Ports: []apiv1.ServicePort{
				{
					Name:     "http",
					Port:     int32(4194), // TODO why not port 80?
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

func (w *worker) GenerateDeployment() (*extensionsv1.Deployment, error) {
	privileged := true

	initContainers, err := ExecTemplate(initContainersWorker, w)
	if err != nil {
		return nil, maskAny(err)
	}
	podAffinity, err := ExecTemplate(podAffinityWorker, w)
	if err != nil {
		return nil, maskAny(err)
	}

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			GenerateName: "worker-",
			Labels: map[string]string{
				"cluster-id": w.Spec.ClusterID,
				"role":       "worker",
				"app":        "k8s-cluster",
			},
			Annotations: map[string]string{
				"pod.beta.kubernetes.io/init-containers": initContainers,
				"scheduler.alpha.kubernetes.io/affinity": podAffinity,
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &w.Spec.Replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					GenerateName: "worker-",
					Labels: map[string]string{
						"cluster-id": w.Spec.ClusterID,
						"role":       w.Spec.ClusterID + "worker",
						"app":        "k8s-cluster",
					},
				},
				Spec: apiv1.PodSpec{
					HostNetwork: true,
					Volumes: []apiv1.Volume{
						{
							Name: "api-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + w.Spec.ClusterID + "/" + w.Spec.ClusterID + "/ssl/worker-1/",
								},
							},
						},
						{
							Name: "calico-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + w.Spec.ClusterID + "/" + w.Spec.ClusterID + "/ssl/worker-1/calico/",
								},
							},
						},
						{
							Name: "etcd-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + w.Spec.ClusterID + "/" + w.Spec.ClusterID + "/ssl/worker-1/etcd/",
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
									Path: "/home/core/vms/" + w.Spec.ClusterID + "-k8s-worker-vm/",
								},
							},
						},
						{
							Name: "ssl",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/ssl/certs",
								},
							},
						},
						{
							Name: "systemctl",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/usr/bin/systemctl",
								},
							},
						},
						{
							Name: "systemd",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/run/systemd",
								},
							},
						},
						{
							Name: "sys-class-net",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/sys/class/net/",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "vm",
							Image: "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-vm:0.9.11",
							Args: []string{
								"worker",
							},
							Env: []apiv1.EnvVar{
								{
									Name: "BRIDGE_NETWORK",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "bridge-network",
										},
									},
								},
								{
									Name: "CUSTOMER_ID",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "customer-id",
										},
									},
								},
								{
									Name:  "DOCKER_EXTRA_ARGS",
									Value: "",
								},
								{
									Name: "G8S_DNS_IP",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
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
												Name: "configmap",
											},
											Key: "g8s-domain",
										},
									},
								},
								{
									Name:  "HOSTNAME",
									Value: w.Spec.ClusterID + "-k8svm-worker-1",
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
									Name: "K8S_INSECURE_PORT",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-insecure-port",
										},
									},
								},
								{
									Name: "K8S_CALICO_MTU",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-calico-mtu",
										},
									},
								},
								{
									Name: "MACHINE_CPU_CORES",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "machine-cpu-cores",
										},
									},
								},
								{
									Name: "K8S_DNS_IP",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-dns-ip",
										},
									},
								},
								{
									Name: "K8S_DOMAIN",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-domain",
										},
									},
								},
								{
									Name: "K8S_ETCD_DOMAIN_NAME",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-etcd-domain-name",
										},
									},
								},
								{
									Name: "K8S_ETCD_PREFIX",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-etcd-prefix",
										},
									},
								},
								{
									Name: "K8S_MASTER_DOMAIN_NAME",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-master-domain-name",
										},
									},
								},
								{
									Name: "K8S_MASTER_PORT",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-master-port",
										},
									},
								},
								{
									Name: "K8S_MASTER_SERVICE_NAME",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-master-service-name",
										},
									},
								},
								{
									Name: "K8S_NETWORK_SETUP_VERSION",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-network-setup-version",
										},
									},
								},
								{
									Name:  "K8S_NODE_LABELS",
									Value: "",
								},
								{
									Name: "K8S_SECURE_PORT",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-secure-port",
										},
									},
								},
								{
									Name: "K8S_VERSION",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-version",
										},
									},
								},
								{
									Name: "MACHINE_MEM",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "machine-mem",
										},
									},
								},
								{
									Name: "REGISTRY",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "registry",
										},
									},
								},
								{
									Name: "DOCKER_EXTRA_ARGS",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "docker-extra-args",
										},
									},
								},
								{
									Name: "K8S_NODE_LABELS",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "k8s-node-labels",
										},
									},
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "certs",
									MountPath: "/etc/kubernetes/ssl/",
								},
								{
									Name:      "images",
									MountPath: "/usr/code/images/",
								},
								{
									Name:      "rootfs",
									MountPath: "/usr/code/rootfs/",
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
