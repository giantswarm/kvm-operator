package flannel

import (
	"fmt"

	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/resources"
)

func (s *Service) newInitContainers(obj interface{}) ([]apiv1.Container, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	privileged := true

	initContainers := []apiv1.Container{
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
					Value: customObject.Spec.Cluster.Flannel.Backend,
				},
				{
					Name:  "BACKEND_VNI", // e.g. 9
					Value: fmt.Sprintf("%d", customObject.Spec.Cluster.Flannel.VNI),
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
					Value: fmt.Sprintf("%d", customObject.Spec.Cluster.Etcd.Port),
				},
				{
					Name:  "NETWORK", // e.g. 10.9.0.0/16
					Value: customObject.Spec.Cluster.Flannel.Network,
				},
				{
					Name:  "NETWORK_BRIDGE_NAME", // e.g. br-h8s2l
					Value: resources.NetworkBridgeName(resources.ClusterID(*customObject)),
				},
				{
					Name:  "SUBNET_LEN", // e.g. 30
					Value: fmt.Sprintf("%d", customObject.Spec.Cluster.Flannel.Client.SubnetLen),
				},
			},
		},
	}

	return initContainers, nil
}
