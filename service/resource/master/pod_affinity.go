package master

import (
	"github.com/giantswarm/kvm-operator/service/resource"
	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (s *Service) newPodAfinity(obj interface{}) (*apiv1.Affinity, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	podAffinity := &apiv1.Affinity{
		PodAntiAffinity: &apiv1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
				{
					LabelSelector: &apismetav1.LabelSelector{
						MatchExpressions: []apismetav1.LabelSelectorRequirement{
							{
								Key:      "app",
								Operator: apismetav1.LabelSelectorOpIn,
								Values: []string{
									"master",
									"worker",
								},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
					Namespaces:  []string{resource.ClusterID(*customObject)},
				},
			},
		},
	}

	return podAffinity, nil
}
