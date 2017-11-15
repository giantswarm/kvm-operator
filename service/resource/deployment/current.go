package deployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for a list of deployments in the Kubernetes API")

	var currentDeployments []*v1beta1.Deployment
	{
		namespace := key.ClusterNamespace(customObject)
		deploymentList, err := r.k8sClient.Extensions().Deployments(namespace).List(apismetav1.ListOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a list of deployments in the Kubernetes API")

			for _, item := range deploymentList.Items {
				d := item
				currentDeployments = append(currentDeployments, &d)
			}
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found a list of %d deployments in the Kubernetes API", len(currentDeployments)))

	return currentDeployments, nil
}
