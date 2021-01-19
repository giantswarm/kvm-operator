package controller

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/wrapper/retryresource"
	"github.com/giantswarm/randomkeys/v2"

	"github.com/giantswarm/kvm-operator/service/controller/resource/deployment"

	"github.com/giantswarm/kvm-operator/service/controller/cloudconfig"
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

	var cloudConfig *cloudconfig.CloudConfig
	{
		c := cloudconfig.Config{
			Logger: config.Logger,

			DockerhubToken: config.DockerhubToken,
			IgnitionPath:   config.IgnitionPath,
			OIDC: cloudconfig.OIDCConfig{
				ClientID:       config.OIDC.ClientID,
				IssuerURL:      config.OIDC.IssuerURL,
				UsernameClaim:  config.OIDC.UsernameClaim,
				UsernamePrefix: config.OIDC.UsernamePrefix,
				GroupsClaim:    config.OIDC.GroupsClaim,
				GroupsPrefix:   config.OIDC.GroupsPrefix,
			},
			RegistryMirrors: config.RegistryMirrors,
			SSOPublicKey:    config.SSOPublicKey,
		}

		cloudConfig, err = cloudconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configmapResource resource.Interface
	{
		c := configmap.Config{
			CertsSearcher:  config.CertsSearcher,
			CloudConfig:    cloudConfig,
			CtrlClient:     config.K8sClient.CtrlClient(),
			KeyWatcher:     randomkeysSearcher,
			Logger:         config.Logger,
			RegistryDomain: config.RegistryDomain,
			DockerhubToken: config.DockerhubToken,
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
