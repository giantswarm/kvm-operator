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

func (s *Service) newDeployment(obj interface{}) (*extensionsv1.Deployment, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	privileged := true
	replicas := int32(1)
	masterNode := customObject.Spec.Cluster.Masters[0]

	deployment := &extensionsv1.Deployment{
		TypeMeta: apiunversioned.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apiv1.ObjectMeta{
			Name: "master",
			Labels: map[string]string{
				"cluster":  resource.ClusterID(*customObject),
				"customer": resource.ClusterCustomer(*customObject),
				"app":      "master",
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
					},
				},
				Spec: apiv1.PodSpec{
					HostNetwork: true,
					Volumes: []apiv1.Volume{
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
									Path: filepath.Join("/home/core/vms/", resource.ClusterID(*customObject), "-k8s-master-vm/"),
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
									Value: fmt.Sprintf("%d", masterNode.CPUs),
								},
								{
									Name: "DISK",
									// TODO this should be configured via clustertpr.Node
									Value: "4G",
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
									Name:  "MEMORY",
									Value: masterNode.Memory,
								},
								{
									Name:  "ROLE",
									Value: "master",
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "images",
									MountPath: "/usr/code/images/",
								},
								{
									Name:      "rootfs",
									MountPath: "/usr/code/rootfs/",
								},
								{
									Name:      "etcd-data",
									MountPath: "/etc/kubernetes/data/etcd/",
								},
								// TODO cloud config has to be written into "/usr/code/cloudconfig/openstack/latest/user_data".
							},
						},
					},
				},
			},
		},
	}

	return deployment, nil
}
