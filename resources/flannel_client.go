package resources

import (
	"encoding/json"
	"fmt"

	"github.com/giantswarm/clusterspec"

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
	clusterspec.Cluster
}

func (f *flannelClient) generateInitFlannelContainers() (string, error) {
	initContainers := []apiv1.Container{
		{
			Name:            "set-network-env",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/set-flannel-network-env",
			ImagePullPolicy: apiv1.PullAlways,
			Command: []string{
				"/bin/bash",
				"-c",
				"/run.sh",
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "CLUSTER_VNI",
					Value: fmt.Sprintf("%d", f.Spec.FlannelConfiguration.ClusterVni),
				},
				{
					Name:  "CLUSTER_NETWORK",
					Value: f.Spec.FlannelConfiguration.ClusterNetwork,
				},
				{
					Name:  "CUSTOMER_ID",
					Value: f.Spec.Customer,
				},
				{
					Name:  "ETCD_PORT",
					Value: f.Spec.GiantnetesConfiguration.EtcdPort,
				},
				{
					Name:  "CLUSTER_ID",
					Value: f.Spec.ClusterId,
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
					Name:  "CLUSTER_BACKEND",
					Value: f.Spec.FlannelConfiguration.ClusterBackend,
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
								Key:      "cluster",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{f.Spec.ClusterId},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
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
		return nil, maskAny(err)
	}

	podAffinity, err := f.generateFlannelPodAffinity()
	if err != nil {
		return nil, maskAny(err)
	}

	flannelClientReplicas := int32(MasterReplicas) + f.Spec.Worker.Replicas

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "flannel-client",
			Labels: map[string]string{
				"cluster":  f.Spec.ClusterId,
				"customer": f.Spec.Customer,
				"app":      "flannel-client",
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: &flannelClientReplicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apiv1.ObjectMeta{
					GenerateName: "flannel-client",
					Labels: map[string]string{
						"cluster":  f.Spec.ClusterId,
						"customer": f.Spec.Customer,
						"app":      "flannel-client",
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
							Name:            "flannel-client",
							Image:           fmt.Sprintf("giantswarm/flannel:%s", f.Spec.FlannelConfiguration.Version),
							ImagePullPolicy: apiv1.PullAlways,
							Command: []string{
								"/bin/sh",
								"-c",
								"/opt/bin/flanneld --remote=$NODE_IP:8889 --public-ip=$NODE_IP --iface=$NODE_IP --networks=$CUSTOMER_ID -v=1",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "CUSTOMER_ID",
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
							Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-network-bridge",
							ImagePullPolicy: apiv1.PullAlways,
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &privileged,
							},
							Command: []string{
								"/bin/sh",
								"-c",
								"while [ ! -f /run/flannel/networks/${CLUSTER_ID}.env ]; do echo 'Waiting for flannel network'; sleep 1; done; /tmp/k8s_network_bridge.sh create ${CLUSTER_ID} ${NETWORK_BRIDGE_NAME} ${NETWORK_INTERFACE} ${HOST_SUBNET_RANGE}",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "CLUSTER_ID",
									Value: f.Spec.ClusterId,
								},
								{
									Name:  "HOST_SUBNET_RANGE",
									Value: f.Spec.GiantnetesConfiguration.HostSubnetRange,
								},
								{
									Name:  "NETWORK_BRIDGE_NAME",
									Value: networkBridgeName(f.Spec.ClusterId),
								},
								{
									Name:  "NETWORK_INTERFACE",
									Value: f.Spec.GiantnetesConfiguration.NetworkInterface,
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
