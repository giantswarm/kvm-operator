package worker

import (
	"github.com/giantswarm/kvm-operator/service/resource"
	"github.com/giantswarm/kvmtpr"
	microerror "github.com/giantswarm/microkit/error"
	"k8s.io/client-go/pkg/api"
	apiunversioned "k8s.io/client-go/pkg/api/unversioned"
)

func (s *Service) newPodAfinity(obj interface{}) (*api.Affinity, error) {
	customObject, ok := obj.(*kvmtpr.CustomObject)
	if !ok {
		return nil, microerror.MaskAnyf(wrongTypeError, "expected '%T', got '%T'", &kvmtpr.CustomObject{}, obj)
	}

	podAffinity := &api.Affinity{
		PodAntiAffinity: &api.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []api.PodAffinityTerm{
				{
					LabelSelector: &apiunversioned.LabelSelector{
						MatchExpressions: []apiunversioned.LabelSelectorRequirement{
							{
								Key:      "app",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{"worker"},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
					Namespaces:  []string{resource.ClusterID(*customObject)},
				},
			},
		},
		PodAffinity: &api.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []api.PodAffinityTerm{
				{
					LabelSelector: &apiunversioned.LabelSelector{
						MatchExpressions: []apiunversioned.LabelSelectorRequirement{
							{
								Key:      "app",
								Operator: apiunversioned.LabelSelectorOpIn,
								Values:   []string{"flannel-client"},
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
