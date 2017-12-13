package deploymentv2

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/giantswarm/kvm-operator/service/keyv2"
)

func newNodeControllerDeployment(customObject v1alpha1.KVMConfig) (*extensionsv1.Deployment, error) {
	replicas := int32(1)

	deployment := &extensionsv1.Deployment{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "deployment",
			APIVersion: "extensions/v1beta",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: keyv2.NodeControllerID,
			Labels: map[string]string{
				"cluster":  keyv2.ClusterID(customObject),
				"customer": keyv2.ClusterCustomer(customObject),
				"app":      keyv2.NodeControllerID,
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Strategy: extensionsv1.DeploymentStrategy{
				Type: extensionsv1.RecreateDeploymentStrategyType,
			},
			Replicas: &replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: apismetav1.ObjectMeta{
					GenerateName: keyv2.NodeControllerID,
					Annotations: map[string]string{
						VersionBundleVersionAnnotation: keyv2.VersionBundleVersion(customObject),
					},
					Labels: map[string]string{
						"app":      keyv2.NodeControllerID,
						"cluster":  keyv2.ClusterID(customObject),
						"customer": keyv2.ClusterCustomer(customObject),
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            keyv2.NodeControllerID,
							Image:           customObject.Spec.KVM.NodeController.Docker.Image,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								fmt.Sprintf("-cluster-api=%s", keyv2.ClusterAPIEndpoint(customObject)),
								fmt.Sprintf("-cluster-id=%s", keyv2.ClusterID(customObject)),
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "PROVIDER_HOST_CLUSTER_NAMESPACE",
									Value: keyv2.ClusterID(customObject),
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
										Path: keyv2.HealthEndpoint,
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
