package deploymentv1

import (
	"fmt"

	"github.com/giantswarm/kvmtpr"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/keyv1"
)

func newNodeControllerDeployment(customObject kvmtpr.CustomObject) (*extensionsv1.Deployment, error) {
	replicas := int32(1)

	deployment := &extensionsv1.Deployment{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: keyv1.NodeControllerID,
			Labels: map[string]string{
				"cluster":  keyv1.ClusterID(customObject),
				"customer": keyv1.ClusterCustomer(customObject),
				"app":      keyv1.NodeControllerID,
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: extensionsv1.RecreateDeploymentStrategyType,
			},
			Replicas: &replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apismetav1.ObjectMeta{
					GenerateName: keyv1.NodeControllerID,
					Annotations: map[string]string{
						VersionBundleVersionAnnotation: keyv1.VersionBundleVersion(customObject),
					},
					Labels: map[string]string{
						"app":      keyv1.NodeControllerID,
						"cluster":  keyv1.ClusterID(customObject),
						"customer": keyv1.ClusterCustomer(customObject),
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            keyv1.NodeControllerID,
							Image:           customObject.Spec.KVM.NodeController.Docker.Image,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								fmt.Sprintf("-cluster-api=%s", keyv1.ClusterAPIEndpoint(customObject)),
								fmt.Sprintf("-cluster-id=%s", keyv1.ClusterID(customObject)),
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "PROVIDER_HOST_CLUSTER_NAMESPACE",
									Value: keyv1.ClusterID(customObject),
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
