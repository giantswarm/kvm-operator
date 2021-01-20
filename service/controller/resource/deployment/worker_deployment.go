package deployment

import (
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func newWorkerDeployment(machine v1alpha2.KVMMachine, cluster v1alpha2.KVMCluster, release releasev1alpha1.Release, dnsServers string, ntpServers string) (*v1.Deployment, error) {
	privileged := true
	replicas := int32(1)
	podDeletionGracePeriod := int64(key.PodDeletionGracePeriod.Seconds())

	containerDistroVersion, err := key.ContainerDistro(release)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	capabilities := machine.Spec.Size

	cpuQuantity, err := key.CPUQuantity(capabilities)
	if err != nil {
		return nil, microerror.Maskf(invalidConfigError, "error creating CPU quantity: %s", err)
	}

	memoryQuantity, err := key.MemoryQuantityWorker(capabilities)
	if err != nil {
		return nil, microerror.Maskf(invalidConfigError, "error creating memory quantity: %s", err)
	}

	deployment := &v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: key.DeploymentName(key.WorkerID, machine.Spec.ProviderID),
			Annotations: map[string]string{
				key.ReleaseVersionAnnotation:       key.ReleaseVersion(cluster),
				key.VersionBundleVersionAnnotation: key.OperatorVersion(cluster),
			},
			Labels: map[string]string{
				key.LabelApp:          key.WorkerID,
				"cluster":             key.ClusterID(&cluster),
				"customer":            key.ClusterCustomer(&cluster),
				key.LabelCluster:      key.ClusterID(&cluster),
				key.LabelOrganization: key.ClusterCustomer(&cluster),
				key.LabelManagedBy:    key.OperatorName,
				"node":                machine.Spec.ProviderID,
			},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					key.LabelApp: key.WorkerID,
					"cluster":    key.ClusterID(&cluster),
					"node":       machine.Spec.ProviderID,
				},
			},
			Strategy: v1.DeploymentStrategy{
				Type: v1.RecreateDeploymentStrategyType,
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						key.AnnotationAPIEndpoint:   key.ClusterAPIEndpoint(cluster),
						key.AnnotationIp:            "",
						key.AnnotationService:       key.WorkerID,
						key.AnnotationPodDrained:    "False",
						key.AnnotationVersionBundle: key.OperatorVersion(cluster),
					},
					Name: key.WorkerID,
					Labels: map[string]string{
						key.LabelApp:          key.WorkerID,
						"cluster":             key.ClusterID(&cluster),
						"customer":            key.ClusterCustomer(&cluster),
						key.LabelCluster:      key.ClusterID(&cluster),
						key.LabelOrganization: key.ClusterCustomer(&cluster),
						"node":                machine.Spec.ProviderID,
						key.PodWatcherLabel:   key.OperatorName,
						label.OperatorVersion: project.Version(),
					},
				},
				Spec: corev1.PodSpec{
					Affinity:    newWorkerPodAfinity(cluster),
					HostNetwork: true,
					NodeSelector: map[string]string{
						"role": key.WorkerID,
					},
					ServiceAccountName:            key.ServiceAccountName(cluster),
					TerminationGracePeriodSeconds: &podDeletionGracePeriod,
					Volumes: []corev1.Volume{
						{
							Name: "cloud-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: key.ConfigMapName(machine, key.WorkerID),
									},
								},
							},
						},
						{
							Name: "images",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: key.FlatcarImageDir,
								},
							},
						},
						{
							Name: "rootfs",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "flannel",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: key.FlannelEnvPathPrefix,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "k8s-endpoint-updater",
							Image:           key.K8SEndpointUpdaterDocker,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"/opt/k8s-endpoint-updater",
								"update",
								"--provider.bridge.name=" + key.NetworkBridgeName(cluster),
								"--service.kubernetes.cluster.namespace=" + key.ClusterNamespace(&cluster),
								"--service.kubernetes.cluster.service=" + key.WorkerID,
								"--service.kubernetes.inCluster=true",
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
							},
						},
						{
							Name:            "k8s-kvm",
							Image:           key.K8SKVMDockerImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Args: []string{
								key.WorkerID,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "CORES",
									Value: fmt.Sprintf("%d", capabilities.CPUs),
								},
								{
									Name:  "FLATCAR_VERSION",
									Value: containerDistroVersion,
								},
								{
									Name:  "FLATCAR_CHANNEL",
									Value: key.FlatcarChannel,
								},
								{
									Name:  "DISK_DOCKER",
									Value: key.DockerVolumeSizeFromNode(capabilities),
								},
								{
									Name:  "DISK_KUBELET",
									Value: key.KubeletVolumeSizeFromNode(capabilities),
								},
								{
									Name:  "DISK_OS",
									Value: key.DefaultOSDiskSize,
								},
								{
									Name:  "DNS_SERVERS",
									Value: dnsServers,
								},
								{
									Name: "HOSTNAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
								{
									Name: "MEMORY",
									// TODO provide memory like disk as float64 and format here.
									Value: capabilities.Memory,
								},
								{
									Name:  "NETWORK_BRIDGE_NAME",
									Value: key.NetworkBridgeName(cluster),
								},
								{
									Name:  "NETWORK_TAP_NAME",
									Value: key.NetworkTapName(cluster),
								},
								{
									Name:  "NTP_SERVERS",
									Value: ntpServers,
								},
								{
									Name:  "ROLE",
									Value: key.WorkerID,
								},
								{
									Name:  "CLOUD_CONFIG_PATH",
									Value: "/cloudconfig/user_data",
								},
							},
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/qemu-shutdown", key.ShutdownDeferrerPollPath(cluster)},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: key.LivenessProbeInitialDelaySeconds,
								TimeoutSeconds:      key.TimeoutSeconds,
								PeriodSeconds:       key.PeriodSeconds,
								FailureThreshold:    key.FailureThreshold,
								SuccessThreshold:    key.SuccessThreshold,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: key.HealthEndpoint,
										Port: intstr.IntOrString{IntVal: key.LivenessPort(cluster)},
										Host: key.ProbeHost,
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: key.ReadinessProbeInitialDelaySeconds,
								TimeoutSeconds:      key.TimeoutSeconds,
								PeriodSeconds:       key.PeriodSeconds,
								FailureThreshold:    key.FailureThreshold,
								SuccessThreshold:    key.SuccessThreshold,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: key.HealthEndpoint,
										Port: intstr.IntOrString{IntVal: key.LivenessPort(cluster)},
										Host: key.ProbeHost,
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU:    cpuQuantity,
									corev1.ResourceMemory: memoryQuantity,
								},
								Limits: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU:    cpuQuantity,
									corev1.ResourceMemory: memoryQuantity,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cloud-config",
									MountPath: "/cloudconfig/",
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
						},
						{
							Name:            "k8s-kvm-health",
							Image:           key.K8SKVMHealthDocker,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "LISTEN_ADDRESS",
									Value: key.HealthListenAddress(cluster),
								},
								{
									Name:  "NETWORK_ENV_FILE_PATH",
									Value: key.NetworkEnvFilePath(cluster),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "flannel",
									MountPath: key.FlannelEnvPathPrefix,
								},
							},
						},
						{
							Name:            "shutdown-deferrer",
							Image:           key.ShutdownDeferrerDocker,
							ImagePullPolicy: corev1.PullAlways,
							Args: []string{
								"daemon",
								"--server.listen.address=" + key.ShutdownDeferrerListenAddress(cluster),
							},
							Env: []corev1.EnvVar{
								{
									Name: key.EnvKeyMyPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: key.EnvKeyMyPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/pre-shutdown-hook", key.ShutdownDeferrerPollPath(cluster)},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: key.LivenessProbeInitialDelaySeconds,
								TimeoutSeconds:      key.TimeoutSeconds,
								PeriodSeconds:       key.PeriodSeconds,
								FailureThreshold:    key.FailureThreshold,
								SuccessThreshold:    key.SuccessThreshold,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: key.HealthEndpoint,
										Port: intstr.IntOrString{IntVal: int32(key.ShutdownDeferrerListenPort(cluster))},
										Host: key.ProbeHost,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	addCoreComponentsAnnotations(deployment, release)

	return deployment, nil
}

func newWorkerPodAfinity(cluster v1alpha2.KVMCluster) *corev1.Affinity {
	podAffinity := &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "app",
								Operator: metav1.LabelSelectorOpIn,
								Values: []string{
									"master",
									"worker",
								},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
					Namespaces: []string{
						key.ClusterID(&cluster),
					},
				},
			},
		},
	}

	return podAffinity
}
