package deployment

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/kvm-operator/service/kvmconfig/v5/key"
)

func newNodeControllerDeployment(customObject v1alpha1.KVMConfig) (*extensionsv1.Deployment, error) {
	replicas := int32(1)

	deployment := &extensionsv1.Deployment{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: key.NodeControllerID,
			Annotations: map[string]string{
				key.VersionBundleVersionAnnotation: key.VersionBundleVersion(customObject),
			},
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
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: key.ServiceAccountName(customObject),
					Containers: []apiv1.Container{
						{
							Name:            key.NodeControllerID,
							Image:           key.NodeControllerDockerImage,
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
							LivenessProbe: &apiv1.Probe{
								InitialDelaySeconds: 15,
								TimeoutSeconds:      1,
								PeriodSeconds:       10,
								FailureThreshold:    3,
								SuccessThreshold:    1,
								Handler: apiv1.Handler{
									HTTPGet: &apiv1.HTTPGetAction{
										Path: key.HealthEndpoint,
										Port: intstr.IntOrString{IntVal: 8080},
									},
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
