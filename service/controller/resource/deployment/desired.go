package deployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/release/v1alpha1"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "reading the release for the deployment")

	var release *releasev1alpha1.Release
	{
		releaseVersion := customResource.Labels[label.ReleaseVersion]
		releaseName := fmt.Sprintf("v%s", releaseVersion)
		release, err = r.g8sClient.ReleaseV1alpha1().Releases().Get(ctx, releaseName, apismetav1.GetOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing the new deployments")

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

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed the %d new deployments", len(deployments)))

	return deployments, nil
}
