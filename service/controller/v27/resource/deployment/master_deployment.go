package deployment

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/kvm-operator/service/controller/v27/key"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newMasterDeployments(customResource v1alpha1.KVMConfig, dnsServers, ntpServers string) ([]*v1.Deployment, error) {
	var deployments []*v1.Deployment

	privileged := true
	replicas := int32(1)
	podDeletionGracePeriod := int64(key.PodDeletionGracePeriod.Seconds())

	for i, masterNode := range customResource.Spec.Cluster.Masters {
		capabilities := customResource.Spec.KVM.Masters[i]

		cpuQuantity, err := key.CPUQuantity(capabilities)
		if err != nil {
			return nil, microerror.Maskf(err, "creating CPU quantity")
		}

		memoryQuantity, err := key.MemoryQuantityMaster(capabilities)
		if err != nil {
			return nil, microerror.Maskf(err, "creating memory quantity")
		}

		storageType := key.StorageType(customResource)

		// During migration, some TPOs do not have storage type set.
		// This specifies a default, until all TPOs have the correct storage type set.
		// tl;dr - this shouldn't be here. If all TPOs have storageType, remove it.
		if storageType == "" {
			storageType = "hostPath"
		}

		var etcdVolume corev1.Volume
		if storageType == "hostPath" {
			etcdVolume = corev1.Volume{
				Name: "etcd-data",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: key.MasterHostPathVolumeDir(key.ClusterID(customResource), key.VMNumber(i)),
					},
				},
			}
		} else if storageType == "persistentVolume" {
			etcdVolume = corev1.Volume{
				Name: "etcd-data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: key.EtcdPVCName(key.ClusterID(customResource), key.VMNumber(i)),
					},
				},
			}
		} else {
			return nil, microerror.Maskf(wrongTypeError, "unknown storageType: '%s'", key.StorageType(customResource))
		}
		deployment := &v1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: key.DeploymentName(key.MasterID, masterNode.ID),
				Annotations: map[string]string{
					key.VersionBundleVersionAnnotation: key.VersionBundleVersion(customResource),
				},
				Labels: map[string]string{
					key.LabelApp:          key.MasterID,
					"cluster":             key.ClusterID(customResource),
					"customer":            key.ClusterCustomer(customResource),
					key.LabelCluster:      key.ClusterID(customResource),
					key.LabelOrganization: key.ClusterCustomer(customResource),
					key.LabelManagedBy:    key.OperatorName,
					"node":                masterNode.ID,
				},
			},
			Spec: v1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						key.LabelApp: key.MasterID,
						"cluster":    key.ClusterID(customResource),
						"node":       masterNode.ID,
					},
				},
				Strategy: v1.DeploymentStrategy{
					Type: v1.RecreateDeploymentStrategyType,
				},
				Replicas: &replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							key.AnnotationAPIEndpoint:   key.ClusterAPIEndpoint(customResource),
							key.AnnotationIp:            "",
							key.AnnotationService:       key.MasterID,
							key.AnnotationPodDrained:    "False",
							key.AnnotationVersionBundle: key.VersionBundleVersion(customResource),
						},
						GenerateName: key.MasterID,
						Labels: map[string]string{
							key.LabelApp:          key.MasterID,
							"cluster":             key.ClusterID(customResource),
							"customer":            key.ClusterCustomer(customResource),
							key.LabelCluster:      key.ClusterID(customResource),
							key.LabelOrganization: key.ClusterCustomer(customResource),
							"node":                masterNode.ID,
							key.PodWatcherLabel:   key.OperatorName,
						},
					},
					Spec: corev1.PodSpec{
						Affinity:    newMasterPodAfinity(customResource),
						HostNetwork: true,
						NodeSelector: map[string]string{
							"role": key.MasterID,
						},
						ServiceAccountName:            key.ServiceAccountName(customResource),
						TerminationGracePeriodSeconds: &podDeletionGracePeriod,
						Volumes: []corev1.Volume{
							{
								Name: "cloud-config",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: key.ConfigMapName(customResource, masterNode, key.MasterID),
										},
									},
								},
							},
							etcdVolume,
							{
								Name: "images",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: key.CoreosImageDir,
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
									"--provider.bridge.name=" + key.NetworkBridgeName(customResource),
									"--service.kubernetes.cluster.namespace=" + key.ClusterNamespace(customResource),
									"--service.kubernetes.cluster.service=" + key.MasterID,
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
									key.MasterID,
								},
								Env: []corev1.EnvVar{
									{
										Name:  "CORES",
										Value: fmt.Sprintf("%d", capabilities.CPUs),
									},
									{
										Name:  "COREOS_VERSION",
										Value: key.CoreosVersion,
									},
									{
										Name:  "DISK_DOCKER",
										Value: key.DefaultDockerDiskSize,
									},
									{
										Name:  "DISK_KUBELET",
										Value: key.DefaultKubeletDiskSize,
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
										Value: key.MasterID,
									},
									{
										Name:  "CLOUD_CONFIG_PATH",
										Value: "/cloudconfig/user_data",
									},
								},
								Lifecycle: &corev1.Lifecycle{
									PreStop: &corev1.Handler{
										Exec: &corev1.ExecAction{
											Command: []string{"/qemu-shutdown", key.ShutdownDeferrerPollPath(customResource)},
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
											Port: intstr.IntOrString{IntVal: key.LivenessPort(customResource)},
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
											Port: intstr.IntOrString{IntVal: key.LivenessPort(customResource)},
											Host: key.ProbeHost,
										},
									},
								},
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
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
										Name:      "etcd-data",
										MountPath: "/etc/kubernetes/data/etcd/",
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
										Name:  "CHECK_K8S_API",
										Value: key.CheckK8sApi,
									},
									{
										Name:  "LISTEN_ADDRESS",
										Value: key.HealthListenAddress(customResource),
									},
									{
										Name:  "NETWORK_ENV_FILE_PATH",
										Value: key.NetworkEnvFilePath(customResource),
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
									"--server.listen.address=" + key.ShutdownDeferrerListenAddress(customResource),
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
											Command: []string{"/pre-shutdown-hook", key.ShutdownDeferrerPollPath(customResource)},
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
