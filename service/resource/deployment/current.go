package deployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for deployments in the Kubernetes API")

	var deployments []*v1beta1.Deployment

	namespace := key.ClusterNamespace(customObject)
	// TODO we have to fetch all deployments within the namespace.
	deploymentNames := key.DeploymentNames(customObject)

	fmt.Printf("%#v\n", deploymentNames)

	for _, name := range deploymentNames {
		manifest, err := r.k8sClient.Extensions().Deployments(namespace).Get(name, apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "did not find a deployment in the Kubernetes API")
			fmt.Printf("current state start\n")
			fmt.Printf("%#v\n", name)
			fmt.Printf("current state end\n")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Log("cluster", key.ClusterID(customObject), "debug", "found a deployment in the Kubernetes API")
			fmt.Printf("current state start\n")
			fmt.Printf("%#v\n", manifest)
			fmt.Printf("current state end\n")

			deployments = append(deployments, manifest)
		}
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", fmt.Sprintf("found %d deployments in the Kubernetes API", len(deployments)))

	return deployments, nil
}
