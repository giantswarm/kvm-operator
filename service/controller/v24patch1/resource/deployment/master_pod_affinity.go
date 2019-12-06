package deployment

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

<<<<<<< HEAD
	"github.com/giantswarm/kvm-operator/service/controller/v24patch1/key"
=======
	"github.com/giantswarm/kvm-operator/service/controller/v24/key"
>>>>>>> c4c6c79d... copy v24 to v24patch1
)

func newMasterPodAfinity(customResource v1alpha1.KVMConfig) *apiv1.Affinity {
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
					Namespaces: []string{
						key.ClusterID(customResource),
					},
				},
			},
		},
	}

	return podAffinity
}
