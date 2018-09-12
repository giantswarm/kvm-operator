package clusterrolebinding

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/rbac/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v14patch1/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for a list of cluster role bindings in the Kubernetes API")

	var currentClusterRoleBinding []*apiv1.ClusterRoleBinding
	{
		clusterRoleBinding, err := r.k8sClient.RbacV1beta1().ClusterRoleBindings().Get(key.ClusterRoleBindingName(customObject), apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find cluster role binding in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found a list of cluster role binding in the Kubernetes API")

			currentClusterRoleBinding = append(currentClusterRoleBinding, clusterRoleBinding)
		}

		clusterRoleBindingPSP, err := r.k8sClient.RbacV1beta1().ClusterRoleBindings().Get(key.ClusterRoleBindingPSPName(customObject), apismetav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find cluster role binding psp in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "found a list of cluster role binding psp in the Kubernetes API")

			currentClusterRoleBinding = append(currentClusterRoleBinding, clusterRoleBindingPSP)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found a list of %d cluster role bindings in the Kubernetes API", len(currentClusterRoleBinding)))

	return currentClusterRoleBinding, nil
}
