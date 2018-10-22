package setup

import (
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/framework/filelogger"
	"github.com/giantswarm/e2e-harness/pkg/framework/resource"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/kvm-operator/integration/env"
)

const (
	organization    = "giantswarm"
	quayAddress     = "https://quay.io"
	tillerNamespace = "kube-system"
)

type Config struct {
	Guest    *framework.Guest
	Host     *framework.Host
	Logger   micrologger.Logger
	Release  *release.Release
	Resource *resource.Resource
}

func NewConfig() (Config, error) {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var apprClient *apprclient.Client
	{
		c := apprclient.Config{
			Logger: logger,

			Address:      quayAddress,
			Organization: organization,
		}

		apprClient, err = apprclient.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var guest *framework.Guest
	{
		c := framework.GuestConfig{
			Logger: logger,

			ClusterID:    env.ClusterID(),
			CommonDomain: env.CommonDomain(),
		}

		guest, err = framework.NewGuest(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var host *framework.Host
	{
		c := framework.HostConfig{
			Logger: logger,

			ClusterID:       env.ClusterID(),
			TargetNamespace: env.TargetNamespace(),
			VaultToken:      env.VaultToken(),
		}

		host, err = framework.NewHost(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var fileLogger *filelogger.FileLogger
	{
		fc := filelogger.Config{
			Backoff:   backoff.NewExponential(backoff.ShortMaxWait, backoff.LongMaxInterval),
			K8sClient: host.K8sClient(),
			Logger:    logger,
		}

		fileLogger, err = filelogger.New(fc)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var helmClient *helmclient.Client
	{
		c := helmclient.Config{
			Logger:          logger,
			K8sClient:       host.K8sClient(),
			RestConfig:      host.RestConfig(),
			TillerNamespace: tillerNamespace,
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var newRelease *release.Release
	{
		c := release.Config{
			ApprClient: apprClient,
			ExtClient:  host.ExtClient(),
			FileLogger: fileLogger,
			G8sClient:  host.G8sClient(),
			HelmClient: helmClient,
			K8sClient:  host.K8sClient(),
			Logger:     logger,

			Namespace: env.TargetNamespace(),
		}

		newRelease, err = release.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var newResource *resource.Resource
	{
		c := resource.Config{
			ApprClient: apprClient,
			HelmClient: helmClient,
			Logger:     logger,

			Namespace: env.TargetNamespace(),
		}

		newResource, err = resource.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	c := Config{
		Guest:    guest,
		Host:     host,
		Logger:   logger,
		Release:  newRelease,
		Resource: newResource,
	}

	return c, nil
}
