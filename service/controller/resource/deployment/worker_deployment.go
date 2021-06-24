package deployment

import (
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func newWorkerDeployments(customResource v1alpha1.KVMConfig, release *releasev1alpha1.Release, dnsServers, ntpServers string) ([]*v1.Deployment, error) {
	var deployments []*v1.Deployment

	privileged := true
	replicas := int32(1)
	podDeletionGracePeriod := int64(key.PodDeletionGracePeriod.Seconds())

	containerDistroVersion, err := key.ContainerDistro(release)
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

		// TODO: https://github.com/giantswarm/giantswarm/issues/17340
		//       https://github.com/giantswarm/kvm-operator/pull/1208#discussion_r636067919
		for _, hostVolume := range capabilities.HostVolumes {
			if hostVolume.MountTag == "" || hostVolume.HostPath == "" {
				return nil, microerror.Maskf(invalidConfigError, "error defining host volume config. both mount tag and host path has to be defined")
			}
		}

		deployment := &v1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: key.DeploymentName(key.WorkerID, workerNode.ID),
				Annotations: map[string]string{
					key.ReleaseVersionAnnotation:       key.ReleaseVersion(customResource),
					key.VersionBundleVersionAnnotation: key.OperatorVersion(customResource),
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
				Selector: &metav1.LabelSelector{
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
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							key.AnnotationAPIEndpoint:   key.ClusterAPIEndpoint(customResource),
							key.AnnotationIp:            "",
							key.AnnotationService:       key.WorkerID,
							key.AnnotationPodDrained:    "False",
							key.AnnotationVersionBundle: key.OperatorVersion(customResource),
						},
						Name: key.WorkerID,
						Labels: map[string]string{
							key.LabelApp:          key.WorkerID,
							"cluster":             key.ClusterID(customResource),
							"customer":            key.ClusterCustomer(customResource),
							key.LabelCluster:      key.ClusterID(customResource),
							key.LabelOrganization: key.ClusterCustomer(customResource),
							"node":                workerNode.ID,
							key.PodWatcherLabel:   key.OperatorName,
							label.OperatorVersion: project.Version(),
						},
					},
					Spec: corev1.PodSpec{
						Affinity: newWorkerPodAfinity(customResource),
						NodeSelector: map[string]string{
							"role": key.WorkerID,
						},
						ServiceAccountName:            key.ServiceAccountName(customResource),
						TerminationGracePeriodSeconds: &podDeletionGracePeriod,
						ReadinessGates: []corev1.PodReadinessGate{
							{
								ConditionType: key.WorkloadClusterNodeReady,
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "ignition",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: key.ConfigMapName(customResource, workerNode, key.WorkerID),
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
										Name:  "CONTAINERVMM_GUEST_CPUS",
										Value: fmt.Sprintf("%d", capabilities.CPUs),
									},
									{
										Name:  "CONTAINERVMM_GUEST_ROOT_DISK_SIZE",
										Value: key.DefaultOSDiskSize,
									},
									{
										Name:  "CONTAINERVMM_FLATCAR_CHANNEL",
										Value: key.FlatcarChannel,
									},
									{
										Name:  "CONTAINERVMM_FLATCAR_VERSION",
										Value: containerDistroVersion,
									},
									{
										Name:  "CONTAINERVMM_FLATCAR_IGNITION_FILE",
										Value: "/var/lib/containervmm/ignition/user_data",
									},
									{
										Name:  "CONTAINERVMM_FLATCAR_IGNITION_FORMAT",
										Value: "base64-compressed",
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
										Name:      "images",
										MountPath: "/var/lib/containervmm/flatcar",
									},
									{
										Name:      "rootfs",
										MountPath: "/var/lib/containervmm/rootfs",
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
						},
					},
				},
			},
		}
		addCoreComponentsAnnotations(deployment, release)
		addHostVolumes(deployment, customResource, i)

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func addHostVolumes(deployment *v1.Deployment, customObject v1alpha1.KVMConfig, workerIndex int) {
	caps := customObject.Spec.KVM.Workers[workerIndex]

	if len(caps.HostVolumes) == 0 {
		return
	}

	for i, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == key.K8SKVMContainerName {
			envVars := []corev1.EnvVar{key.HostVolumesToEnvVar(caps.HostVolumes)}
			container.Env = append(container.Env, envVars...)

			volumeMounts := key.HostVolumesToVolumeMounts(caps.HostVolumes)
			container.VolumeMounts = append(container.VolumeMounts, volumeMounts...)

			deployment.Spec.Template.Spec.Containers[i] = container
		}
	}

	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, key.HostVolumesToVolumes(customObject, workerIndex)...)
}
