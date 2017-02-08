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
			Name:            "k8s-network-config",
			Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-network-config:7e6b155f78ce00b2193c3015863e1994e97ed4b5",
			ImagePullPolicy: apiv1.PullAlways,
			Env: []apiv1.EnvVar{
				{
					Name:  "BACKEND_TYPE", // e.g. vxlan
					Value: f.Spec.FlannelConfiguration.ClusterBackend,
				},
				{
					Name:  "BACKEND_VNI", // e.g. 9
					Value: fmt.Sprintf("%d", f.Spec.FlannelConfiguration.ClusterVni),
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
					Name:  "ETCD_PORT",
					Value: f.Spec.GiantnetesConfiguration.EtcdPort,
				},
				{
					Name:  "NETWORK", // e.g. 10.9.0.0/16
					Value: f.Spec.FlannelConfiguration.ClusterNetwork,
				},
				{
					Name:  "NETWORK_BRIDGE_NAME", // e.g. br-h8s2l
					Value: networkBridgeName(f.Spec.ClusterId),
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
								Key:      "app",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{"flannel-client"},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
					Namespaces:  []string{f.Spec.ClusterId},
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
							Image:           fmt.Sprintf("quay.io/coreos/flannel:%s", f.Spec.FlannelConfiguration.Version),
							ImagePullPolicy: apiv1.PullAlways,
							Command: []string{
								"/bin/sh",
								"-c",
								"/opt/bin/flanneld --remote=$NODE_IP:8889 --public-ip=$NODE_IP --iface=$NODE_IP --networks=$NETWORK_BRIDGE_NAME -v=1",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "NETWORK_BRIDGE_NAME",
									Value: networkBridgeName(f.Spec.ClusterId),
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
							Name:            "k8s-network-bridge",
							Image:           "leaseweb-registry.private.giantswarm.io/giantswarm/k8s-network-bridge:test-move-1",
							ImagePullPolicy: apiv1.PullAlways,
							Command: []string{
								"/bin/sh",
								"-c",
								"while [ ! -f ${NETWORK_ENV_FILE_PATH} ]; do echo 'Waiting for ${NETWORK_ENV_FILE_PATH} to be created'; sleep 1; done; /docker-entrypoint.sh create ${NETWORK_ENV_FILE_PATH} ${NETWORK_BRIDGE_NAME} ${NETWORK_INTERFACE_NAME} ${HOST_SUBNET_RANGE}",
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "NETWORK_ENV_FILE_PATH",
									Value: networkEnvFilePath(f.Spec.ClusterId),
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
									Name:  "NETWORK_INTERFACE_NAME",
									Value: f.Spec.GiantnetesConfiguration.NetworkInterface,
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &privileged,
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
