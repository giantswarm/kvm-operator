package deployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "reading the release for the deployment")

	var release *releasev1alpha1.Release
	{
		releaseVersion := customResource.Labels[label.ReleaseVersion]
		var release releasev1alpha1.Release
		err = r.ctrlClient.Get(ctx, client.ObjectKey{
			Name: fmt.Sprintf("v%s", releaseVersion),
		}, &release)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	r.logger.Debugf(ctx, "computing the new deployments")

	var deployments []*v1.Deployment

	{
		masterDeployments, err := newMasterDeployments(customResource, release, r.dnsServers, r.ntpServers)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		deployments = append(deployments, masterDeployments...)

		workerDeployments, err := newWorkerDeployments(customResource, release, r.dnsServers, r.ntpServers)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		deployments = append(deployments, workerDeployments...)
	}

	r.logger.Debugf(ctx, "computed the %d new deployments", len(deployments))

	return deployments, nil
}
