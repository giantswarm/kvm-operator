package controller

import (
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"

	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const podAffinityFlannelClient string = `
{
	"podAntiAffinity": {
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
	 }
 }`

const initContainerSetNetworkEnv string = `[
	{
		"name": "set-network-env",
		"image": "hectorj2f/set-flannel-network-env",
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
				"name": "CLUSTER_NETWORK",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "{{.ClusterID}}-configmap",
						"key": "cluster-network"
					}
				}
			},
			{
				"name": "CLUSTER_VNI",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "{{.ClusterID}}-configmap",
						"key": "cluster-vni"
					}
				}
			},
			{
				"name": "CLUSTER_BACKEND",
				"valueFrom": {
					"configMapKeyRef": {
						"name": "{{.ClusterID}}-configmap",
						"key": "cluster-backend"
					}
				}
			},
			{
				"name": "ETCD_ENDPOINT",
				"valueFrom": {
					"fieldRef": {
						"fieldPath": "spec.nodeName"
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
			}
		],
		"command": [
			"/bin/bash",
			"-c",
			"/run.sh"
		]
	}
]`

type FlannelClient interface {
	ClusterObj
}

type flannelClient struct {
	ClusterConfig
}


func (f *flannelClient) Create() error {
	privileged := true

	initContainers, err := ExecTemplate(initContainerSetNetworkEnv, f)
	if err != nil {
		return maskAny(err)
	}

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			GenerateName: f.ClusterID+"-flannel-client",
			Labels: map[string]string{
				"cluster-id": f.ClusterID,
				"role": f.ClusterID+"-flannel-client",
				"app": f.ClusterID+"-k8s-cluster",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &f.Replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					GenerateName: f.ClusterID+"flannel-client",
					Labels: map[string]string{
						"cluster-id": f.ClusterID,
						"role": f.ClusterID+"-flannel-client",
						"app": f.ClusterID+"-k8s-cluster",
					},
					Annotations: map[string]string{
						"seccomp.security.alpha.kubernetes.io/pod": "unconfined",
						"pod.beta.kubernetes.io/init-containers": initContainers,
						"scheduler.alpha.kubernetes.io/affinity": podAffinityFlannelClient,
					},
				},
				Spec: apiv1.PodSpec{
					HostNetwork: true,
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: "cgroup",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/sys/fs/cgroup",
								},
							},
						},
						apiv1.Volume{
							Name: "dbus",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/var/run/dbus",
								},
							},
						},
						apiv1.Volume{
							Name: "environment",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/environment",
								},
							},
						},
						apiv1.Volume{
							Name: "etc-systemd",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/systemd/",
								},
							},
						},
						apiv1.Volume{
							Name: "flannel",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/run/flannel",
								},
							},
						},
						apiv1.Volume{
							Name: "ssl",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/ssl/certs",
								},
							},
						},
						apiv1.Volume{
							Name: "systemctl",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/usr/bin/systemctl",
								},
							},
						},
						apiv1.Volume{
							Name: "systemd",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/run/systemd",
								},
							},
						},
						apiv1.Volume{
							Name: "sys-class-net",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/sys/class/net/",
								},
							},
						},
					},
					Containers: []apiv1.Container{
					apiv1.Container{
						Name:  "flannel-client",
						Image: "giantswarm/flannel:v0.6.2",
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
						ImagePullPolicy: apiv1.PullAlways,
						SecurityContext: &apiv1.SecurityContext{
							Privileged: &privileged,
						},
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
					},
				},
				},
			},
		},
	}

	if 	_, err := f.KubernetesClient.Extensions().Deployments(f.Namespace).Create(deployment); err != nil {
		return maskAny(err)
	}

	return nil
}


func (f *flannelClient)  Delete() error {
	return nil
}
