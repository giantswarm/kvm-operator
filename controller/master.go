package controller

import (
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"

	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
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
								"values": ["{{.ClusterID}}-worker"]
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
						"name": "{{.ClusterID}}-configmap",
						"key": "cluster-id"
					}
				}
			},
			{
				"name": "ETCD_PORT",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "{{.ClusterID}}-configmap",
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

type Master interface {
	ClusterObj
}

type master struct{
	ClusterConfig
}

// TODO retry operations in case they fail
func (m *master) Create() error {
	if err := m.CreateKubernetesMasterDeployment(); err != nil {
		return maskAny(err)
	}

	if err := m.CreateKubernetesMasterService(); err != nil {
		return maskAny(err)
	}

	return nil
}

func (m *master) CreateKubernetesMasterService() error {
	endpointMasterEtcd := &extensionsv1.Ingress{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			GenerateName: "etcd",
			Labels: map[string]string{
				"cluster-id": m.ClusterID,
				"role": m.ClusterID+"-master",
				"app": m.ClusterID+"-k8s-cluster",
			},
			Namespace: m.Namespace,
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: m.ClusterID+"-master",
				ServicePort: intstr.FromInt(2379),
			},
		},
	}

	_, err := m.KubernetesClient.Extensions().Ingresses(m.Namespace).Create(endpointMasterEtcd)
	if err != nil {
		return maskAny(err)
	}

	endpointMasterAPIHTTP := &extensionsv1.Ingress{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			GenerateName: "api",
			Labels: map[string]string{
				"cluster-id": m.ClusterID,
				"role": m.ClusterID+"-master",
				"app": m.ClusterID+"-k8s-cluster",
			},
			Namespace: m.Namespace,
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: m.ClusterID+"-master",
				ServicePort: intstr.FromInt(8080),
			},
		},
	}

	_, err = m.KubernetesClient.Extensions().Ingresses(m.Namespace).Create(endpointMasterAPIHTTP)
	if err != nil {
		return maskAny(err)
	}

	endpointMasterAPIHTTPS := &extensionsv1.Ingress{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			GenerateName: "api-https",
			Labels: map[string]string{
				"cluster-id": m.ClusterID,
				"role": m.ClusterID+"-master",
				"app": m.ClusterID+"-k8s-cluster",
			},
			Namespace: m.Namespace,
		},
		Spec: extensionsv1.IngressSpec{
			Backend: &extensionsv1.IngressBackend{
				ServiceName: m.ClusterID+"-master",
				ServicePort: intstr.FromInt(6443),
			},
		},
	}

	_, err = m.KubernetesClient.Extensions().Ingresses(m.Namespace).Create(endpointMasterAPIHTTPS)
	if err != nil {
		return maskAny(err)
	}

	service := &apiv1.Service{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: apiv1.ObjectMeta{
			GenerateName: m.ClusterID+"-k8s-master",
			Labels: map[string]string{
				"cluster-id": m.ClusterID,
				"role": m.ClusterID+"-master",
				"app": m.ClusterID+"-k8s-cluster",
			},
			Namespace: m.Namespace,
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

	_, err = m.KubernetesClient.Core().Services(m.Namespace).Create(service)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (m *master) CreateKubernetesMasterDeployment() error {
	privileged := true

	initContainers, err := ExecTemplate(initContainersMaster, m)
	if err != nil {
		return maskAny(err)
	}

	podAffinity, err := ExecTemplate(podAffinityMaster, m)
	if err != nil {
		return maskAny(err)
	}

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			GenerateName: "master-",
			Labels: map[string]string{
				"cluster-id": m.ClusterID,
				"role": m.ClusterID+"-master",
				"app": m.ClusterID+"-k8s-cluster",
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
			Replicas: &m.Replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					GenerateName: m.ClusterID+"-master",
					Labels: map[string]string{
						"cluster-id": m.ClusterID,
						"role": m.ClusterID+"-master",
						"app": m.ClusterID+"-k8s-cluster",
					},
				},
				Spec: apiv1.PodSpec{
					HostNetwork: true,
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: "etcd-data",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/home/core/"+m.ClusterID+"-k8s-master-vm/",
								},
							},
						},
						apiv1.Volume{
							Name: "customer-dir",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/"+m.ClusterID+"/"+m.ClusterID+"/",
								},
							},
						},
						apiv1.Volume{
							Name: "api-secrets",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/"+m.ClusterID+"/"+m.ClusterID+"/secrets",
								},
							},
						},
						apiv1.Volume{
							Name: "calico-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/"+m.ClusterID+"/"+m.ClusterID+"/ssl/master/calico/",
								},
							},
						},
						apiv1.Volume{
							Name: "etcd-certs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/kubernetes/"+m.ClusterID+"/"+m.ClusterID+"/ssl/master/etcd/",
								},
							},
						},
						apiv1.Volume{
							Name: "images",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/home/core/images/",
								},
							},
						},
						apiv1.Volume{
							Name: "rootfs",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/home/core/vms/"+m.ClusterID+"-k8s-master-vm/",
								},
							},
						},
						apiv1.Volume{
							Name: "ssl",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/ssl/certs/ca-certificates.crt",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						apiv1.Container{
							Name:  "vm",
							Image: "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-vm:0.9.11",
							Args: []string{
								"master",
							},
							Env: []apiv1.EnvVar{
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
									Name:  "DOCKER_EXTRA_ARGS",
									Value: "",
								},
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
									Name:  "HOSTNAME",
									Value: m.ClusterID+"-k8svm-master",
								},
								apiv1.EnvVar{
									Name: "HOST_PUBLIC_IP",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
								apiv1.EnvVar{
									Name:  "IP_BRIDGE",
									Value: "",
								},
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
									Name:  "K8S_NODE_LABELS",
									Value: "",
								},
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.EnvVar{
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
								apiv1.VolumeMount{
									Name:      "certs",
									MountPath: "/etc/kubernetes/ssl/",
								},
								apiv1.VolumeMount{
									Name:      "images",
									MountPath: "/usr/code/images/",
								},
								apiv1.VolumeMount{
									Name:      "rootfs",
									MountPath: "/usr/code/rootfs/",
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &privileged,
							},
						},
						apiv1.Container{
							Name:  "flannel",
							Image: "leaseweb-registry.private.giantswarm.io/giantswarm/flannel:v0.6.2",
							Command: []string{
								"/bin/sh",
								"-c",
								"/opt/bin/flanneld --remote=$NODE_IP:8889 --public-ip=$NODE_IP --iface=$NODE_IP --networks=$CUSTOMER_ID -v=1",
							},
							Env: []apiv1.EnvVar{
								apiv1.EnvVar{
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
								apiv1.EnvVar{
									Name: "NODE_IP",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      "flannel",
									MountPath: "/run/flannel",
								},
								apiv1.VolumeMount{
									Name:      "ssl",
									MountPath: "/etc/ssl/certs",
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &privileged,
							},
						},
						apiv1.Container{
							Name:  "create-bridge",
							Image: "hectorj2f/k8s-network-bridge", // TODO: Sort this image out (giantswarm, needs tag)
							Command: []string{
								"/bin/sh",
								"-c",
								"while [ ! -f /run/flannel/networks/${CUSTOMER_ID}.env ]; do echo 'Waiting for flannel network'; sleep 1; done; /tmp/k8s_network_bridge.sh create ${CUSTOMER_ID} br${CUSTOMER_ID} ${NETWORK_INTERFACE} ${HOST_SUBNET_RANGE}",
							},
							Env: []apiv1.EnvVar{
								apiv1.EnvVar{
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
								apiv1.EnvVar{
									Name: "HOST_SUBNET_RANGE",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "host-subnet-range",
										},
									},
								},
								apiv1.EnvVar{
									Name: "NETWORK_INTERFACE",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "configmap",
											},
											Key: "network-interface",
										},
									},
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      "cgroup",
									MountPath: "/sys/fs/cgroup",
								},
								apiv1.VolumeMount{
									Name:      "dbus",
									MountPath: "/var/run/dbus",
								},
								apiv1.VolumeMount{
									Name:      "environment",
									MountPath: "/etc/environment",
								},
								apiv1.VolumeMount{
									Name:      "etc-systemd",
									MountPath: "/etc/systemd/",
								},
								apiv1.VolumeMount{
									Name:      "flannel",
									MountPath: "/run/flannel",
								},
								apiv1.VolumeMount{
									Name:      "systemctl",
									MountPath: "/usr/bin/systemctl",
								},
								apiv1.VolumeMount{
									Name:      "systemd",
									MountPath: "/run/systemd",
								},
								apiv1.VolumeMount{
									Name:      "sys-class-net",
									MountPath: "/sys/class/net/",
								},
							},
							ImagePullPolicy: apiv1.PullIfNotPresent,
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &privileged,
							},
						},
					},
				},
			},
		},
	}

	_, err = m.KubernetesClient.Extensions().Deployments(m.Namespace).Create(deployment)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (m *master)  Delete() error {
	return nil
}
