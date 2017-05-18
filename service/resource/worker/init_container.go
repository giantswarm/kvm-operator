package worker

import (
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
			Name:            "k8s-endpoint-updater",
			Image:           customObject.Spec.KVM.EndpointUpdater.Docker.Image,
			ImagePullPolicy: apiv1.PullIfNotPresent,
			Command: []string{
				"/bin/sh",
				"-c",
				"/opt/k8s-endpoint-updater update --provider.bridge.name=${NETWORK_BRIDGE_NAME} --provider.kind=bridge --service.kubernetes.address=\"\" --service.kubernetes.cluster.namespace=${POD_NAMESPACE} --service.kubernetes.cluster.service=worker --service.kubernetes.inCluster=true --updater.pod.names=${POD_NAME}",
			},
			SecurityContext: &apiv1.SecurityContext{
				Privileged: &privileged,
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "NETWORK_BRIDGE_NAME",
					Value: resource.NetworkBridgeName(resource.ClusterID(*customObject)),
				},
				{
					Name: "POD_NAME",
					ValueFrom: &apiv1.EnvVarSource{
						FieldRef: &apiv1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.name",
						},
					},
				},
				{
					Name: "POD_NAMESPACE",
					ValueFrom: &apiv1.EnvVarSource{
						FieldRef: &apiv1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.namespace",
						},
					},
				},
			},
		},
	}

	return initContainers, nil
}
