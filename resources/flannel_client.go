package resources

import (
	"encoding/json"
	"fmt"

	"k8s.io/client-go/pkg/api"
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
)

type FlannelClient interface {
	ClusterObj
}

type flannelClient struct {
	Cluster
}

func (f *flannelClient) generateInitFlannelContainers() (string, error) {
	initContainers := []apiv1.Container{
		{
			Name:  "set-network-env",
			Image: "leaseweb-registry.private.giantswarm.io/giantswarm/set-flannel-network-env",
			Command: []string{
				"/bin/bash",
				"-c",
				"/run.sh",
			},
			Env: []apiv1.EnvVar{
				{
					Name: "CLUSTER_VNI",
					Value: fmt.Sprintf("%d", f.Spec.ClusterVNI),
				},
				{
					Name: "CLUSTER_NETWORK",
					Value: f.Spec.ClusterNetwork,
				},
				{
					Name: "CUSTOMER_ID",
					Value: f.Spec.Customer,
				},
				{
					Name: "ETCD_PORT",
					ValueFrom: &apiv1.EnvVarSource{
						ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
							LocalObjectReference: apiv1.LocalObjectReference{
								Name: GiantnetesConfigMapName,
							},
							Key: "etcd-port",
						},
					},
				},
				{
					Name: "CLUSTER_ID",
					Value: f.Spec.ClusterID,
				},
				{
					Name: "ETCD_ENDPOINT",
					ValueFrom: &apiv1.EnvVarSource{
						FieldRef: &apiv1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "spec.nodeName",
						},
					},
				},
				{
					Name: "CLUSTER_BACKEND",
					Value: f.Spec.ClusterBackend,
				},
			},
		},
	}

	bytesInitContainers, err := json.Marshal(initContainers)
	if err != nil {
		return "", maskAny(err)
	}
	return string(bytesInitContainers), nil
}

func (f *flannelClient) generateFlannelPodAffinity() (string, error) {
	podAntiAffinity := &api.Affinity{
		PodAntiAffinity: &api.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []api.PodAffinityTerm{
				{
					LabelSelector: &apiunversioned.LabelSelector{
						MatchExpressions: []apiunversioned.LabelSelectorRequirement{
							{
								Key:      "role",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{f.Spec.ClusterID + "-flannel-client"},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		},
	}

	bytesPodAffinity, err := json.Marshal(podAntiAffinity)
	if err != nil {
		return "", maskAny(err)
	}

	return string(bytesPodAffinity), nil
}

func (f *flannelClient) GenerateResources() ([]runtime.Object, error) {
	privileged := true

	initContainers, err := f.generateInitFlannelContainers()
	if err != nil {
		return []runtime.Object{}, maskAny(err)
	}

	podAffinity, err := f.generateFlannelPodAffinity()
	if err != nil {
		return []runtime.Object{}, maskAny(err)
	}

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: f.Spec.ClusterID + "-flannel-client",
			Labels: map[string]string{
				"cluster-id": f.Spec.ClusterID,
				"role":       f.Spec.ClusterID + "-flannel-client",
				"app":        f.Spec.ClusterID + "-k8s-cluster",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &f.Spec.NumNodes,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					GenerateName: f.Spec.ClusterID + "flannel-client",
					Labels: map[string]string{
						"cluster-id": f.Spec.ClusterID,
						"role":       f.Spec.ClusterID + "-flannel-client",
						"app":        f.Spec.ClusterID + "-k8s-cluster",
					},
					Annotations: map[string]string{
						"seccomp.security.alpha.kubernetes.io/pod": "unconfined",
						"pod.beta.kubernetes.io/init-containers":   string(initContainers),
						"scheduler.alpha.kubernetes.io/affinity":   string(podAffinity),
					},
				},
				Spec: apiv1.PodSpec{
					HostNetwork: true,
					Volumes: []apiv1.Volume{
						{
							Name: "cgroup",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/sys/fs/cgroup",
								},
							},
						},
						{
							Name: "dbus",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/var/run/dbus",
								},
							},
						},
						{
							Name: "environment",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/environment",
								},
							},
						},
						{
							Name: "etc-systemd",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/etc/systemd/",
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
							Name:  "flannel-client",
							Image: "giantswarm/flannel:v0.6.2",
							Command: []string{
								"/bin/sh",
								"-c",
								"/opt/bin/flanneld --remote=$NODE_IP:8889 --public-ip=$NODE_IP --iface=$NODE_IP --networks=$CUSTOMER_ID -v=1",
							},
							Env: []apiv1.EnvVar{
								{
									Name: "CUSTOMER_ID",
									Value: f.Spec.Customer,
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
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "flannel",
									MountPath: "/run/flannel",
								},
								{
									Name:      "ssl",
									MountPath: "/etc/ssl/certs",
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &privileged,
							},
						},
						{
							Name:            "create-bridge",
							Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-network-bridge", // TODO: Sort this image out (giantswarm, needs tag)
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
								{
									Name: "CUSTOMER_ID",
									Value: f.Spec.Customer,
								},
								{
									Name: "HOST_SUBNET_RANGE",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: GiantnetesConfigMapName,
											},
											Key: "host-subnet-range",
										},
									},
								},
								{
									Name: "NETWORK_INTERFACE",
									ValueFrom: &apiv1.EnvVarSource{
										ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: GiantnetesConfigMapName,
											},
											Key: "network-interface",
										},
									},
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "cgroup",
									MountPath: "/sys/fs/cgroup",
								},
								{
									Name:      "dbus",
									MountPath: "/var/run/dbus",
								},
								{
									Name:      "environment",
									MountPath: "/etc/environment",
								},
								{
									Name:      "etc-systemd",
									MountPath: "/etc/systemd/",
								},
								{
									Name:      "flannel",
									MountPath: "/run/flannel",
								},
								{
									Name:      "systemctl",
									MountPath: "/usr/bin/systemctl",
								},
								{
									Name:      "systemd",
									MountPath: "/run/systemd",
								},
								{
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

	objects := append([]runtime.Object{}, deployment)

	return objects, nil
}
