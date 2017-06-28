package operator

import (
	"time"

	"github.com/giantswarm/kvmtpr"
	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var (
	ClusterThirdPartyResource = v1beta1.ThirdPartyResource{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "ThirdPartyResource",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: kvmtpr.Name,
		},
		Description: "A specification of a Kubernetes cluster",
		Versions: []v1beta1.APIVersion{
			v1beta1.APIVersion{
				Name: "v1",
			},
		},
	}

	clusterResourceCreation = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_third_party_resource_creation_milliseconds",
		Help: "Time taken to create cluster third party resource, in milliseconds",
	})
)

func init() {
	prometheus.MustRegister(clusterResourceCreation)
}

// createClusterResource creates the 'cluster' ThirdPartyResource,
// if it does not exist already.
func (s *Service) createClusterResource() error {
	tprClient := s.kubernetesClient.Extensions().ThirdPartyResources()

	start := time.Now()

	s.logger.Log("debug", "creating cluster resource")
	var err error
	if _, err = tprClient.Create(&ClusterThirdPartyResource); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	if apierrors.IsAlreadyExists(err) {
		s.logger.Log("debug", "cluster resource already exists")
	} else {
		s.logger.Log("debug", "cluster resource created")
	}

	clusterResourceCreation.Set(float64(time.Since(start) / time.Millisecond))

	return nil
}
