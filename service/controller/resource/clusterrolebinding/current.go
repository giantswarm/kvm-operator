package clusterrolebinding

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/resourcecanceledcontext"
	apiv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if key.IsDeleted(customObject) {
		r.logger.Debugf(ctx, "redirecting responsibility of deletion of cluster role bindings to namespace termination")
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.Debugf(ctx, "canceling resource")

		return nil, nil
	}

	r.logger.Debugf(ctx, "looking for a list of cluster role bindings in the Kubernetes API")

	var currentClusterRoleBinding []*apiv1.ClusterRoleBinding
	{
		var clusterRoleBinding apiv1.ClusterRoleBinding
		err := r.ctrlClient.Get(ctx, client.ObjectKey{Name: key.ClusterRoleBindingName(customObject)}, &clusterRoleBinding)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find cluster role binding in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found a list of cluster role binding in the Kubernetes API")

			currentClusterRoleBinding = append(currentClusterRoleBinding, &clusterRoleBinding)
		}

		var clusterRoleBindingPSP apiv1.ClusterRoleBinding
		err = r.ctrlClient.Get(ctx, client.ObjectKey{Name: key.ClusterRoleBindingPSPName(customObject)}, &clusterRoleBinding)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find cluster role binding psp in the Kubernetes API")
			// fall through
		} else if err != nil {
			return nil, microerror.Mask(err)
		} else {
			r.logger.Debugf(ctx, "found a list of cluster role binding psp in the Kubernetes API")

			currentClusterRoleBinding = append(currentClusterRoleBinding, &clusterRoleBindingPSP)
		}
	}

	r.logger.Debugf(ctx, "found a list of %d cluster role bindings in the Kubernetes API", len(currentClusterRoleBinding))

	return currentClusterRoleBinding, nil
}
