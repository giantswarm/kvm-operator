package resources

import (
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"

	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/intstr"
)

const podAffinityMaster string = `{
		"podAffinity": {
			"requiredDuringSchedulingIgnoredDuringExecution": [
				{
					"labelSelector": {
						"matchExpressions": [
							{
								"key": "role",
								"operator": "In",
								"values": ["{{.Spec.ClusterID}}-flannel-client"]
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
								"values": ["{{.Spec.ClusterID}}-worker"]
							}
						]
					},
					"topologyKey": "kubernetes.io/hostname"
			 }
			]
		 }
	 }`

const initContainersMaster string = `[
{
		"name": "set-iptables",
		"image": "hectorj2f/alpine-bash-iptables",
		"securityContext": {
			"privileged": true
		},
		"restartPolicy": "Never",
		"imagePullPolicy": "Always",
		"env": [
			{
				"name": "CLUSTER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "cluster-id"
					}
				}
			},
			{
				"name": "ETCD_PORT",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "etcd-port"
					}
				}
			},
			{
				"name": "NODE_IP",
				"valueFrom": {
					"fieldRef": {
						"fieldPath": "spec.nodeName"
					}
				}
			}
		],
		"command": [
			"/bin/sh",
			"-c",
			"/sbin/iptables -I INPUT -p tcp --match multiport --dports $ETCD_PORT -d ${NODE_IP} -i br${CLUSTER_ID} -j ACCEPT"
		]
	},
	{
		"name": "generate-bridgeip-configmap",
		"image": "hectorj2f/generate-bridge-ip-configmap",
		"securityContext": {
			"privileged": true
		},
		"imagePullPolicy": "Always",
		"volumeMounts": [
			{
				"name": "customer-dir",
				"mountPath": "/tmp/"
			}
		],
		"env": [
			{
				"name": "CLUSTER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "cluster-id"
					}
				}
			},
			{
				"name": "CUSTOMER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "customer-id"
					}
				}
			},
			{
				"name": "SUFFIX_CONFIGMAP",
				"value": "master-vm"
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
		"imagePullPolicy": "IfNotPresent",
		"volumeMounts": [
			{
				"name": "customer-dir",
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
				"value": "/tmp/bridge-ip-configmap-master-vm.json"
			}
		],
		"command": [
			"/bin/sh",
			"-c",
			"while [ ! -f /tmp/bridge-ip-configmap-master-vm.json ]; do echo -; sleep 1; done; /usr/bin/kubectl --server=${G8S_MASTER_HOST}:${G8S_MASTER_PORT} replace --force -f ${BRIDGE_IP_CONFIGMAP_PATH}"
		]
	},
	 {
		"name": "k8s-master-api-token",
		"securityContext": {
			"privileged": true
		},
		"image": "hectorj2f/alpine-openssl",
		"imagePullPolicy": "IfNotPresent",
		"volumeMounts": [
			{
				"name": "api-secrets",
				"mountPath": "/etc/kubernetes/secrets"
			},
			{
				"name": "ssl",
				"mountPath": "/etc/ssl/certs/ca-certificates.crt"
			}
		],
		"command": [
			"/bin/sh",
			"-c",
			"/usr/bin/test ! -f /etc/kubernetes/secrets/token_sign_key.pem  && /usr/bin/openssl genrsa -out /etc/kubernetes/secrets/token_sign_key.pem 2048 && /bin/echo \"Generated new token sign key.\" || /bin/echo \"Token sign key already exists, skipping.\""
		],
		"env": [
			{
				"name": "VAULT_ADDR",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "vault-addr"
					}
				}
			},
			{
				"name": "VAULT_TOKEN",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "vault-token"
					}
				}
			},
			{
				"name": "CLUSTER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "cluster-id"
					}
				}
			},
			{
				"name": "CUSTOMER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "customer-id"
					}
				}
			},
			{
				"name": "G8S_API_IP",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "g8s-api-ip"
					}
				}
			},
			{
				"name": "K8S_API_ALT_NAMES",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "k8s-api-alt-names"
					}
				}
			}
		]
	},
	{
		"name": "k8s-master-api-certs",
		"securityContext": {
			"privileged": true
		},
		"image": "giantswarm/certctl:0.5.0",
		"imagePullPolicy": "IfNotPresent",
		"volumeMounts": [
			{
				"name": "api-certs",
				"mountPath": "/etc/kubernetes/ssl/master/"
			},
			{
				"name": "ssl",
				"mountPath": "/etc/ssl/certs/ca-certificates.crt"
			}
		],
		"command": [
			"/bin/sh",
			"-c",
			"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=api.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/master/apiserver.pem --key-file=/etc/kubernetes/ssl/master/apiserver-key.pem --ca-file=/etc/kubernetes/ssl/master/apiserver-ca.pem --alt-names=$K8S_MASTER_SERVICE_NAME,$K8S_API_ALT_NAMES --ip-sans=$G8S_API_IP"
		],
		"env": [
			{
				"name": "VAULT_ADDR",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "vault-addr"
					}
				}
			},
			{
				"name": "VAULT_TOKEN",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "vault-token"
					}
				}
			},
			{
				"name": "CLUSTER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "cluster-id"
					}
				}
			},
			{
				"name": "CUSTOMER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "customer-id"
					}
				}
			},
			{
				"name": "G8S_API_IP",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "g8s-api-ip"
					}
				}
			},
			{
				"name": "K8S_API_ALT_NAMES",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "k8s-api-alt-names"
					}
				}
			},
			{
				"name": "K8S_MASTER_SERVICE_NAME",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "k8s-master-service-name"
					}
				}
			}
		]
	},
	{
		"name": "k8s-master-etcd-certs",
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
			"/opt/certctl issue --vault-addr=$VAULT_ADDR --vault-token=$VAULT_TOKEN --cluster-id=$CLUSTER_ID --common-name=etcd.$CUSTOMER_ID.g8s.fra-1.giantswarm.io --ttl=720h --crt-file=/etc/kubernetes/ssl/etcd/server.pem --key-file=/etc/kubernetes/ssl/etcd/server-key.pem --ca-file=/etc/kubernetes/ssl/etcd/server-ca.pem --alt-names=$K8S_MASTER_SERVICE_NAME"
		],
		"env": [
			{
				"name": "VAULT_ADDR",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "vault-addr"
					}
				}
			},
			{
				"name": "VAULT_TOKEN",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "vault-token"
					}
				}
			},
			{
				"name": "CLUSTER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "cluster-id"
					}
				}
			},
			{
				"name": "CUSTOMER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "customer-id"
					}
				}
			},
			{
				"name": "K8S_MASTER_SERVICE_NAME",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "k8s-master-service-name"
					}
				}
			}
		]
	},
	{
		"name": "k8s-master-calico-certs",
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
						"name": "configmap",
						"key": "vault-addr"
					}
				}
			},
			{
				"name": "VAULT_TOKEN",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "vault-token"
					}
				}
			},
			{
				"name": "CLUSTER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "cluster-id"
					}
				}
			},
			{
				"name": "CUSTOMER_ID",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "customer-id"
					}
				}
			},
			{
				"name": "K8S_MASTER_SERVICE_NAME",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "configmap",
						"key": "k8s-master-service-name"
					}
				}
			}
		]
	}
]`

type Master interface {
	ClusterObj
}

type master struct {
	Cluster
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
			GenerateName: "etcd",
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
			GenerateName: "api",
			Labels: map[string]string{
				"cluster-id": m.Spec.ClusterID,
				"role":       m.Spec.ClusterID + "-master",
				"app":        m.Spec.ClusterID + "-k8s-cluster",
			},
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: m.Spec.ClusterID + "-master",
				ServicePort: intstr.FromInt(8080),
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
			GenerateName: "api-https",
			Labels: map[string]string{
				"cluster-id": m.Spec.ClusterID,
				"role":       m.Spec.ClusterID + "-master",
				"app":        m.Spec.ClusterID + "-k8s-cluster",
			},
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: m.Spec.ClusterID + "-master",
				ServicePort: intstr.FromInt(6443),
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
			GenerateName: m.Spec.ClusterID + "-k8s-master",
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
					Port:     int32(8080),
					Protocol: "TCP",
				},
				{
					Name:     "etcd",
					Port:     int32(2379),
					Protocol: "TCP",
				},
				{
					Name:     "api-https",
					Port:     int32(6443),
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

	initContainers, err := ExecTemplate(initContainersMaster, m)
	if err != nil {
		return &extensionsv1.Deployment{}, maskAny(err)
	}

	podAffinity, err := ExecTemplate(podAffinityMaster, m)
	if err != nil {
		return &extensionsv1.Deployment{}, maskAny(err)
	}

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
			Annotations: map[string]string{
				"pod.beta.kubernetes.io/init-containers": initContainers,
				"scheduler.alpha.kubernetes.io/affinity": podAffinity,
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &m.Spec.Replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					GenerateName: m.Spec.ClusterID + "-master",
					Labels: map[string]string{
						"cluster-id": m.Spec.ClusterID,
						"role":       m.Spec.ClusterID + "-master",
						"app":        m.Spec.ClusterID + "-k8s-cluster",
					},
				},
				Spec: apiv1.PodSpec{
					HostNetwork: true,
					Volumes: []apiv1.Volume{
						{
							Name: "etcd-data",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/home/core/" + m.Spec.ClusterID + "-k8s-master-vm/",
								},
							},
						},
						{
							Name: "customer-dir",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + m.Spec.ClusterID + "/" + m.Spec.ClusterID + "/",
								},
							},
						},
						{
							Name: "api-secrets",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/" + m.Spec.ClusterID + "/" + m.Spec.ClusterID + "/secrets",
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
									Path: "/etc/ssl/certs",
								},
							},
						},
						{
							Name: "certs",
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
							Name:  "vm",
							Image: "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-vm:0.9.13",
							Args: []string{
								"master",
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
									Value: m.Spec.ClusterID + "-k8svm-master",
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
						{
							Name:  "watch-master-vm-service",
							Image: "hectorj2f/watch-master-vm-service",
							Command: []string{
								"/bin/sh",
								"-c",
								"/run.sh",
							},
							Env: []apiv1.EnvVar{
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
