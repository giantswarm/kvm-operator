package deployment

import (
	"fmt"

	"github.com/giantswarm/kvmtpr"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func newNodeControllerDeployment(customObject kvmtpr.CustomObject) (*extensionsv1.Deployment, error) {
	replicas := int32(1)

	deployment := &extensionsv1.Deployment{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: key.NodeControllerID,
			Labels: map[string]string{
				"cluster":  key.ClusterID(customObject),
				"customer": key.ClusterCustomer(customObject),
				"app":      key.NodeControllerID,
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: extensionsv1.RecreateDeploymentStrategyType,
			},
			Replicas: &replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apismetav1.ObjectMeta{
					GenerateName: key.NodeControllerID,
					Labels: map[string]string{
						"app":      key.NodeControllerID,
						"cluster":  key.ClusterID(customObject),
						"customer": key.ClusterCustomer(customObject),
					},
					Annotations: map[string]string{},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            key.NodeControllerID,
							Image:           customObject.Spec.KVM.NodeController.Docker.Image,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								fmt.Sprintf("-cluster-api=%s", key.ClusterAPIEndpoint(customObject)),
								fmt.Sprintf("-cluster-id=%s", key.ClusterID(customObject)),
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "PROVIDER_HOST_CLUSTER_NAMESPACE",
									Value: key.ClusterID(customObject),
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment, nil
}
