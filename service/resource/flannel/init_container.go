package flannel

import (
	"fmt"

	"github.com/giantswarm/kvmtpr"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/resource"
)

func (s *Service) newInitContainers(customObject *kvmtpr.CustomObject) []apiv1.Container {
	privileged := true

	return []apiv1.Container{
		{
			Name:            "k8s-network-config",
			Image:           customObject.Spec.KVM.Network.Config.Docker.Image,
			ImagePullPolicy: apiv1.PullAlways,
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "BACKEND_TYPE", // e.g. vxlan
					Value: customObject.Spec.KVM.Flannel.Backend,
				},
				{
					Name:  "BACKEND_VNI", // e.g. 9
					Value: fmt.Sprintf("%d", customObject.Spec.KVM.Flannel.VNI),
				},
				{
					Name:  "ETCD_ENDPOINT",
					Value: "https://127.0.0.1:2379",
				},
				{
					Name:  "NETWORK", // e.g. 10.9.0.0/16
					Value: customObject.Spec.KVM.Flannel.Network,
				},
				{
					Name:  "NETWORK_BRIDGE_NAME", // e.g. br-h8s2l
					Value: resource.NetworkBridgeName(resource.ClusterID(*customObject)),
				},
				{
					Name:  "SUBNET_LEN", // e.g. 30
					Value: fmt.Sprintf("%d", customObject.Spec.KVM.Flannel.Client.SubnetLen),
				},
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "etcd-certs",
					MountPath: "/etc/kubernetes/ssl/etcd/",
				},
			},
		},
	}
}
