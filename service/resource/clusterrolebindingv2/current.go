package clusterrolebindingv2

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/rbac/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	r.logger.LogCtx(ctx, "debug", "looking for a list of cluster role bindings in the Kubernetes API")

	var currentClusterRoleBinding []*apiv1.ClusterRoleBinding
	{
		clusterRoleBindingList, err := r.k8sClient.RbacV1beta1().ClusterRoleBindings().List(apismetav1.ListOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "debug", "did not find any cluster role bindings in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "debug", "found a list of cluster role bindings in the Kubernetes API")

			for _, item := range clusterRoleBindingList.Items {
				c := item
				currentClusterRoleBinding = append(currentClusterRoleBinding, &c)
			}
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("found a list of %d cluster role bindings in the Kubernetes API", len(currentClusterRoleBinding)))

	return currentClusterRoleBinding, nil
}
