package deployment

import (
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/kvm-operator/v4/pkg/label"
	"github.com/giantswarm/kvm-operator/v4/pkg/project"
	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func newMasterDeployments(customResource v1alpha1.KVMConfig, release *releasev1alpha1.Release, dnsServers, ntpServers string) ([]*v1.Deployment, error) {
	var deployments []*v1.Deployment

	privileged := true
	replicas := int32(1)
	podDeletionGracePeriod := int64(key.PodDeletionGracePeriod.Seconds())

	containerDistroVersion, err := key.ContainerDistro(release)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for i, masterNode := range customResource.Spec.Cluster.Masters {
		capabilities := customResource.Spec.KVM.Masters[i]

		cpuQuantity, err := key.CPUQuantity(capabilities)
		if err != nil {
			return nil, microerror.Maskf(invalidConfigError, "error creating CPU quantity: %s", err)
		}

		memoryQuantity, err := key.MemoryQuantityMaster(capabilities)
		if err != nil {
			return nil, microerror.Maskf(invalidConfigError, "error creating memory quantity: %s", err)
		}

		storageType := key.EtcdStorageType(customResource)

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
			return nil, microerror.Maskf(wrongTypeError, "unknown storageType: '%s'", key.EtcdStorageType(customResource))
		}
		deployment := &v1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: key.DeploymentName(key.MasterID, masterNode.ID),
				Annotations: map[string]string{
					key.ReleaseVersionAnnotation:       key.ReleaseVersion(customResource),
					key.VersionBundleVersionAnnotation: key.OperatorVersion(customResource),
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
							key.AnnotationService:       key.MasterID,
							key.AnnotationPodDrained:    "False",
							key.AnnotationVersionBundle: key.OperatorVersion(customResource),
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
							label.OperatorVersion: project.Version(),
						},
					},
					Spec: corev1.PodSpec{
						Affinity: newMasterPodAfinity(customResource),
						NodeSelector: map[string]string{
							"role": key.MasterID,
						},
						ServiceAccountName:            key.ServiceAccountName(customResource),
						TerminationGracePeriodSeconds: &podDeletionGracePeriod,
						Volumes: []corev1.Volume{
							etcdVolume,
							{
								Name: "ignition",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: key.ConfigMapName(customResource, masterNode, key.MasterID),
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
								Name: "disks",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
							{
								Name: "dev-kvm",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/dev/kvm",
									},
								},
							},
							{
								Name: "dev-net-tun",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/dev/net/tun",
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:            key.K8SKVMContainerName,
								Image:           key.K8SKVMDockerImage,
								ImagePullPolicy: corev1.PullIfNotPresent,
								SecurityContext: &corev1.SecurityContext{
									Privileged: &privileged,
									Capabilities: &corev1.Capabilities{
										Add: []corev1.Capability{
											"NET_ADMIN",
										},
									},
								},
								Env: []corev1.EnvVar{
									{
										Name:  "CONTAINERVMM_GUEST_CPUS",
										Value: fmt.Sprintf("%d", capabilities.CPUs),
									},
									{
										Name:  "CONTAINERVMM_FLATCAR_VERSION",
										Value: containerDistroVersion,
									},
									{
										Name:  "CONTAINERVMM_FLATCAR_CHANNEL",
										Value: key.FlatcarChannel,
									},
									{
										Name:  "CONTAINERVMM_GUEST_ROOT_DISK_SIZE",
										Value: key.DefaultOSDiskSize,
									},
									{
										Name:  "CONTAINERVMM_GUEST_DNS_SERVERS",
										Value: dnsServers,
									},
									{
										Name: "CONTAINERVMM_GUEST_NAME",
										ValueFrom: &corev1.EnvVarSource{
											FieldRef: &corev1.ObjectFieldSelector{
												APIVersion: "v1",
												FieldPath:  "metadata.name",
											},
										},
									},
									{
										Name: "CONTAINERVMM_GUEST_MEMORY",
										// TODO provide memory like disk as float64 and format here.
										Value: capabilities.Memory,
									},
									{
										Name:  "CONTAINERVMM_GUEST_NTP_SERVERS",
										Value: ntpServers,
									},
									{
										Name:  "CONTAINERVMM_FLATCAR_IGNITION_FILE",
										Value: "/var/lib/containervmm/ignition/ignition",
									},
									{
										Name: "CONTAINERVMM_GUEST_HOST_VOLUMES",
										Value: key.HostVolumesEnvVarValue([]v1alpha1.KVMConfigSpecKVMNodeHostVolumes{
											{
												MountTag: "etcdshare",
												HostPath: "/var/lib/containervmm/etcd",
											},
										}),
									},
									{
										Name: "CONTAINERVMM_GUEST_ADDITIONAL_DISKS",
										Value: strings.Join([]string{
											strings.Join([]string{"dockerfs", key.DefaultDockerDiskSize}, ":"),
											strings.Join([]string{"kubeletfs", key.DefaultKubeletDiskSize}, ":"),
										}, ","),
									},
								},
								Lifecycle: &corev1.Lifecycle{
									PreStop: &corev1.Handler{
										Exec: &corev1.ExecAction{
											Command: []string{"/usr/local/bin/qemu-shutdown", key.ShutdownDeferrerPollPath(customResource)},
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
										Name:      "ignition",
										MountPath: "/var/lib/containervmm/ignition",
									},
									{
										Name:      "etcd",
										MountPath: "/var/lib/containervmm/etcd",
									},
									{
										Name:      "images",
										MountPath: "/var/lib/containervmm/images",
									},
									{
										Name:      "disks",
										MountPath: "/var/lib/containervmm/disks",
									},
									{
										Name:      "dev-kvm",
										MountPath: "/dev/kvm",
									},
									{
										Name:      "dev-net-tun",
										MountPath: "/dev/net/tun",
									},
								},
							},
							{
								Name:            "k8s-kvm-health",
								Image:           key.K8SKVMHealthDocker,
								ImagePullPolicy: corev1.PullIfNotPresent,
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
										Name: "IP_ADDRESS",
										ValueFrom: &corev1.EnvVarSource{
											FieldRef: &corev1.ObjectFieldSelector{
												FieldPath: "status.podIP",
											},
										},
									},
								},
							},
							{
								Name:            "shutdown-deferrer",
								Image:           key.ShutdownDeferrerDocker,
								ImagePullPolicy: corev1.PullIfNotPresent,
								Args: []string{
									"daemon",
									"--server.listen.address=" + key.ShutdownDeferrerListenAddress(customResource),
								},
							},
						},
					},
				},
			},
		}
		addCoreComponentsAnnotations(deployment, release)

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}
