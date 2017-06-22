package master

import (
	"fmt"
	"path/filepath"

	"github.com/giantswarm/kvm-operator/service/resource"
	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func (s *Service) newDeployments(obj interface{}) ([]*extensionsv1.Deployment, error) {
	var deployments []*extensionsv1.Deployment

	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	privileged := true
	replicas := int32(1)

	for i, masterNode := range customObject.Spec.Cluster.Masters {
		capabilities := customObject.Spec.KVM.Masters[i]

		deployment := &extensionsv1.Deployment{
			TypeMeta: apiunversioned.TypeMeta{
				Kind:       "deployment",
				APIVersion: "extensions/v1beta",
			},
			ObjectMeta: apiv1.ObjectMeta{
				GenerateName: "master-",
				Labels: map[string]string{
					"cluster":  resource.ClusterID(*customObject),
					"customer": resource.ClusterCustomer(*customObject),
					"app":      "master",
					"node":     masterNode.ID,
				},
			},
			Spec: extensionsv1.DeploymentSpec{
				Strategy: extensionsv1.DeploymentStrategy{
					Type: extensionsv1.RecreateDeploymentStrategyType,
				},
				Replicas: &replicas,
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: apiv1.ObjectMeta{
						GenerateName: "master",
						Labels: map[string]string{
							"cluster":  resource.ClusterID(*customObject),
							"customer": resource.ClusterCustomer(*customObject),
							"app":      "master",
							"node":     masterNode.ID,
						},
						Annotations: map[string]string{},
					},
					Spec: apiv1.PodSpec{
						HostNetwork: true,
						Volumes: []apiv1.Volume{
							{
								Name: "cloud-config",
								VolumeSource: apiv1.VolumeSource{
									ConfigMap: &apiv1.ConfigMapVolumeSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: resource.ConfigMapName(*customObject, masterNode, "master"),
										},
									},
								},
							},
							{
								Name: "etcd-data",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: filepath.Join("/home/core/", resource.ClusterID(*customObject), "-k8s-master-vm/"),
									},
								},
							},
							{
								Name: "images",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: "/home/core/images/",
									},
								},
							},
							{
								Name: "rootfs",
								VolumeSource: apiv1.VolumeSource{
									HostPath: &apiv1.HostPathVolumeSource{
										Path: filepath.Join("/home/core/vms", resource.ClusterID(*customObject), masterNode.ID),
									},
								},
							},
						},
						Containers: []apiv1.Container{
							{
								Name:            "k8s-kvm",
								Image:           customObject.Spec.KVM.K8sKVM.Docker.Image,
								ImagePullPolicy: apiv1.PullIfNotPresent,
								SecurityContext: &apiv1.SecurityContext{
									Privileged: &privileged,
								},
								Args: []string{
									"master",
								},
								Env: []apiv1.EnvVar{
									{
										Name:  "CORES",
										Value: fmt.Sprintf("%d", capabilities.CPUs),
									},
									{
										Name:  "DISK",
										Value: fmt.Sprintf("%.0fG", capabilities.Disk),
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
										Name:  "NETWORK_BRIDGE_NAME",
										Value: resource.NetworkBridgeName(resource.ClusterID(*customObject)),
									},
									{
										Name: "MEMORY",
										// TODO provide memory like disk as float64 and format here.
										Value: capabilities.Memory,
									},
									{
										Name:  "ROLE",
										Value: "master",
									},
								},
								VolumeMounts: []apiv1.VolumeMount{
									{
										Name:      "cloud-config",
										MountPath: "/usr/code/cloudconfig/openstack/latest/",
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
						},
					},
				},
			},
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}
