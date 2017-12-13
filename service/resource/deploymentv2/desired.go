package deploymentv2

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"k8s.io/api/extensions/v1beta1"

	"github.com/giantswarm/kvm-operator/service/keyv2"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "computing the new deployments")

	var deployments []*v1beta1.Deployment

	{
		masterDeployments, err := newMasterDeployments(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		deployments = append(deployments, masterDeployments...)

		workerDeployments, err := newWorkerDeployments(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		deployments = append(deployments, workerDeployments...)

		if keyv2.HasNodeController(customObject) {
			nodeControllerDeployment, err := newNodeControllerDeployment(customObject)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			deployments = append(deployments, nodeControllerDeployment)
		}
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("computed the %d new deployments", len(deployments)))

	return deployments, nil
}
