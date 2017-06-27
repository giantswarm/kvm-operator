package master

import (
	"fmt"

	"github.com/giantswarm/kvm-operator/service/resource"
	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (s *Service) newInitContainers(obj interface{}) ([]apiv1.Container, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	privileged := true

	initContainers := []apiv1.Container{
		{
			Name:            "set-iptables",
			Image:           customObject.Spec.KVM.Network.IPTables.Docker.Image,
			ImagePullPolicy: apiv1.PullIfNotPresent,
			Command: []string{
				"/bin/sh",
				"-c",
				"/sbin/iptables -I INPUT -p tcp --match multiport --dports ${ETCD_PORT} -d ${NODE_IP} -i ${NETWORK_BRIDGE_NAME} -j ACCEPT",
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "ETCD_PORT",
					Value: fmt.Sprintf("%d", customObject.Spec.Cluster.Etcd.Port),
				},
				{
					Name:  "NETWORK_BRIDGE_NAME",
					Value: resource.NetworkBridgeName(resource.ClusterID(*customObject)),
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
		},
	}

	return initContainers, nil
}
