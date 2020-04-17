package deployment

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/kvm-operator/service/controller/key"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newWorkerDeployments(customResource v1alpha1.KVMConfig, release *releasev1alpha1.Release, dnsServers, ntpServers string) ([]*v1.Deployment, error) {
	var deployments []*v1.Deployment

	privileged := true
	replicas := int32(1)
	podDeletionGracePeriod := int64(key.PodDeletionGracePeriod.Seconds())

	containerDistro, err := key.ContainerDistro(release)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for i, workerNode := range customResource.Spec.Cluster.Workers {
		capabilities := customResource.Spec.KVM.Workers[i]

		cpuQuantity, err := key.CPUQuantity(capabilities)
		if err != nil {
			return nil, microerror.Maskf(invalidConfigError, "error creating CPU quantity: %s", err)
		}

		memoryQuantity, err := key.MemoryQuantityWorker(capabilities)
		if err != nil {
			return nil, microerror.Maskf(invalidConfigError, "error creating memory quantity: %s", err)
		}

		deployment := &v1.Deployment{
			TypeMeta: apismetav1.TypeMeta{
				Kind:       "deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: apismetav1.ObjectMeta{
				Name: key.DeploymentName(key.WorkerID, workerNode.ID),
				Annotations: map[string]string{
					key.VersionBundleVersionAnnotation: key.VersionBundleVersion(customResource),
				},
				Labels: map[string]string{
					key.LabelApp:          key.WorkerID,
					"cluster":             key.ClusterID(customResource),
					"customer":            key.ClusterCustomer(customResource),
					key.LabelCluster:      key.ClusterID(customResource),
					key.LabelOrganization: key.ClusterCustomer(customResource),
					key.LabelManagedBy:    key.OperatorName,
					"node":                workerNode.ID,
				},
			},
			Spec: v1.DeploymentSpec{
				Selector: &apismetav1.LabelSelector{
					MatchLabels: map[string]string{
						key.LabelApp: key.WorkerID,
						"cluster":    key.ClusterID(customResource),
						"node":       workerNode.ID,
					},
				},
				Strategy: v1.DeploymentStrategy{
					Type: v1.RecreateDeploymentStrategyType,
				},
				Replicas: &replicas,
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: apismetav1.ObjectMeta{
						Annotations: map[string]string{
							key.AnnotationAPIEndpoint:   key.ClusterAPIEndpoint(customResource),
							key.AnnotationIp:            "",
							key.AnnotationService:       key.WorkerID,
							key.AnnotationPodDrained:    "False",
							key.AnnotationVersionBundle: key.VersionBundleVersion(customResource),
						},
						Name: key.WorkerID,
						Labels: map[string]string{
							key.LabelApp:          key.WorkerID,
							"cluster":             key.ClusterID(customResource),
							"customer":            key.ClusterCustomer(customResource),
							key.LabelCluster:      key.ClusterID(customResource),
							key.LabelOrganization: key.ClusterCustomer(customResource),
							"node":                workerNode.ID,
							key.PodWatcherLabel:   "kvm-operator",
						},
					},
					Spec: apiv1.PodSpec{
						Affinity:    newWorkerPodAfinity(customResource),
						HostNetwork: true,
						NodeSelector: map[string]string{
							"role": key.WorkerID,
						},
						ServiceAccountName:            key.ServiceAccountName(customResource),
						TerminationGracePeriodSeconds: &podDeletionGracePeriod,
						Volumes: []apiv1.Volume{
							{
								Name: "cloud-config",
								VolumeSource: apiv1.VolumeSource{
									ConfigMap: &apiv1.ConfigMapVolumeSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: key.ConfigMapName(customResource, workerNode, key.WorkerID),
										},
									},
								},
							},
							{
								Name: "images",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: key.FlatcarImageDir,
									},
								},
							},
							{
								Name: "rootfs",
								VolumeSource: apiv1.VolumeSource{
									EmptyDir: &apiv1.EmptyDirVolumeSource{},
								},
							},
							{
								Name: "flannel",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: key.FlannelEnvPathPrefix,
									},
								},
							},
						},
						Containers: []apiv1.Container{
							{
								Name:            "k8s-endpoint-updater",
								Image:           key.K8SEndpointUpdaterDocker,
								ImagePullPolicy: apiv1.PullIfNotPresent,
								Command: []string{
									"/opt/k8s-endpoint-updater",
									"update",
									"--provider.bridge.name=" + key.NetworkBridgeName(customResource),
									"--service.kubernetes.cluster.namespace=" + key.ClusterNamespace(customResource),
									"--service.kubernetes.cluster.service=" + key.WorkerID,
									"--service.kubernetes.inCluster=true",
								},
								SecurityContext: &apiv1.SecurityContext{
									Privileged: &privileged,
								},
								Env: []apiv1.EnvVar{
									{
										Name: "POD_NAME",
										ValueFrom: &apiv1.EnvVarSource{
											FieldRef: &apiv1.ObjectFieldSelector{
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
								ImagePullPolicy: apiv1.PullIfNotPresent,
								SecurityContext: &apiv1.SecurityContext{
									Privileged: &privileged,
								},
								Args: []string{
									key.WorkerID,
								},
								Env: []apiv1.EnvVar{
									{
										Name:  "CORES",
										Value: fmt.Sprintf("%d", capabilities.CPUs),
									},
									{
										Name:  "FLATCAR_VERSION",
										Value: containerDistro,
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
										ValueFrom: &apiv1.EnvVarSource{
											FieldRef: &apiv1.ObjectFieldSelector{
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
										Value: key.NetworkBridgeName(customResource),
									},
									{
										Name:  "NETWORK_TAP_NAME",
										Value: key.NetworkTapName(customResource),
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
								Lifecycle: &apiv1.Lifecycle{
									PreStop: &apiv1.Handler{
										Exec: &apiv1.ExecAction{
											Command: []string{"/qemu-shutdown", key.ShutdownDeferrerPollPath(customResource)},
										},
									},
								},
								LivenessProbe: &apiv1.Probe{
									InitialDelaySeconds: key.LivenessProbeInitialDelaySeconds,
									TimeoutSeconds:      key.TimeoutSeconds,
									PeriodSeconds:       key.PeriodSeconds,
									FailureThreshold:    key.FailureThreshold,
									SuccessThreshold:    key.SuccessThreshold,
									Handler: apiv1.Handler{
										HTTPGet: &apiv1.HTTPGetAction{
											Path: key.HealthEndpoint,
											Port: intstr.IntOrString{IntVal: key.LivenessPort(customResource)},
											Host: key.ProbeHost,
										},
									},
								},
								ReadinessProbe: &apiv1.Probe{
									InitialDelaySeconds: key.ReadinessProbeInitialDelaySeconds,
									TimeoutSeconds:      key.TimeoutSeconds,
									PeriodSeconds:       key.PeriodSeconds,
									FailureThreshold:    key.FailureThreshold,
									SuccessThreshold:    key.SuccessThreshold,
									Handler: apiv1.Handler{
										HTTPGet: &apiv1.HTTPGetAction{
											Path: key.HealthEndpoint,
											Port: intstr.IntOrString{IntVal: key.LivenessPort(customResource)},
											Host: key.ProbeHost,
										},
									},
								},
								Resources: apiv1.ResourceRequirements{
									Requests: map[apiv1.ResourceName]resource.Quantity{
										apiv1.ResourceCPU:    cpuQuantity,
										apiv1.ResourceMemory: memoryQuantity,
									},
									Limits: map[apiv1.ResourceName]resource.Quantity{
										apiv1.ResourceCPU:    cpuQuantity,
										apiv1.ResourceMemory: memoryQuantity,
									},
								},
								VolumeMounts: []apiv1.VolumeMount{
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
								ImagePullPolicy: apiv1.PullAlways,
								Env: []apiv1.EnvVar{
									{
										Name:  "LISTEN_ADDRESS",
										Value: key.HealthListenAddress(customResource),
									},
									{
										Name:  "NETWORK_ENV_FILE_PATH",
										Value: key.NetworkEnvFilePath(customResource),
									},
								},
								SecurityContext: &apiv1.SecurityContext{
									Privileged: &privileged,
								},
								VolumeMounts: []apiv1.VolumeMount{
									{
										Name:      "flannel",
										MountPath: key.FlannelEnvPathPrefix,
									},
								},
							},
							{
								Name:            "shutdown-deferrer",
								Image:           key.ShutdownDeferrerDocker,
								ImagePullPolicy: apiv1.PullAlways,
								Args: []string{
									"daemon",
									"--server.listen.address=" + key.ShutdownDeferrerListenAddress(customResource),
								},
								Env: []apiv1.EnvVar{
									{
										Name: key.EnvKeyMyPodName,
										ValueFrom: &apiv1.EnvVarSource{
											FieldRef: &apiv1.ObjectFieldSelector{
												FieldPath: "metadata.name",
											},
										},
									},
									{
										Name: key.EnvKeyMyPodNamespace,
										ValueFrom: &apiv1.EnvVarSource{
											FieldRef: &apiv1.ObjectFieldSelector{
												FieldPath: "metadata.namespace",
											},
										},
									},
								},
								Lifecycle: &apiv1.Lifecycle{
									PreStop: &apiv1.Handler{
										Exec: &apiv1.ExecAction{
											Command: []string{"/pre-shutdown-hook", key.ShutdownDeferrerPollPath(customResource)},
										},
									},
								},
								LivenessProbe: &apiv1.Probe{
									InitialDelaySeconds: key.LivenessProbeInitialDelaySeconds,
									TimeoutSeconds:      key.TimeoutSeconds,
									PeriodSeconds:       key.PeriodSeconds,
									FailureThreshold:    key.FailureThreshold,
									SuccessThreshold:    key.SuccessThreshold,
									Handler: apiv1.Handler{
										HTTPGet: &apiv1.HTTPGetAction{
											Path: key.HealthEndpoint,
											Port: intstr.IntOrString{IntVal: int32(key.ShutdownDeferrerListenPort(customResource))},
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

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}
