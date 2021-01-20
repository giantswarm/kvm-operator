package controller

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/randomkeys/v2"

	"github.com/giantswarm/kvm-operator/service/controller/resource/deployment"

	"github.com/giantswarm/kvm-operator/service/controller/resource/configmap"
)

func newMachineResources(config MachineConfig) ([]resource.Interface, error) {
	var err error

	var randomkeysSearcher randomkeys.Interface
	{
		c := randomkeys.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		randomkeysSearcher, err = randomkeys.NewSearcher(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configmapResource resource.Interface
	{
		c := configmap.Config{
			CertsSearcher:   config.CertsSearcher,
			K8sClient:       config.K8sClient,
			KeyWatcher:      randomkeysSearcher,
			Logger:          config.Logger,
			DockerhubToken:  config.DockerhubToken,
			RegistryDomain:  config.RegistryDomain,
			IgnitionPath:    config.IgnitionPath,
			RegistryMirrors: config.RegistryMirrors,
		}

		configmapResource, err = configmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deploymentResource resource.Interface
	{
		c := deployment.Config{
			CtrlClient:    config.K8sClient.CtrlClient(),
			Logger:        config.Logger,
			TenantCluster: config.TenantCluster,

			DNSServers: config.DNSServers,
			NTPServers: config.NTPServers,
		}

		deploymentResource, err = deployment.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		configmapResource,
		deploymentResource,
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
