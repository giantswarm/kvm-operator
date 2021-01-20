package controller

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/statusresource/v3"

	"github.com/giantswarm/kvm-operator/service/controller/key"
	"github.com/giantswarm/kvm-operator/service/controller/resource/cluster"
	"github.com/giantswarm/kvm-operator/service/controller/resource/namespace"
	"github.com/giantswarm/kvm-operator/service/controller/resource/nodeindexstatus"
)

func newTransitionResources(config TransitionConfig) ([]resource.Interface, error) {
	var err error

	var statusResource resource.Interface
	{
		c := statusresource.ResourceConfig{
			ClusterEndpointFunc:      key.ToClusterEndpoint,
			ClusterIDFunc:            key.ToClusterID,
			ClusterStatusFunc:        key.ToClusterStatus,
			NodeCountFunc:            key.ToNodeCount,
			Logger:                   config.Logger,
			RESTClient:               config.K8sClient.G8sClient().ProviderV1alpha1().RESTClient(),
			TenantCluster:            config.TenantCluster,
			VersionBundleVersionFunc: key.ToOperatorVersion,
		}

		statusResource, err = statusresource.NewResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var nodeIndexStatusResource resource.Interface
	{
		c := nodeindexstatus.Config{
			G8sClient: config.K8sClient.G8sClient(),
			Logger:    config.Logger,
		}

		nodeIndexStatusResource, err = nodeindexstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var namespaceResource resource.Interface
	{
		c := namespace.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		ops, err := namespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		namespaceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterResource resource.Interface
	{
		c := cluster.Config{
			CtrlClient: config.K8sClient.CtrlClient(),
			Logger:     config.Logger,
		}

		clusterResource, err = cluster.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		statusResource,
		nodeIndexStatusResource,
		namespaceResource,
		clusterResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}
