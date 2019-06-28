package setup

import (
	"context"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/crdstorage"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/framework/resource"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/kvm-operator/integration/env"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/retrystorage"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace       = "giantswarm"
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
	Storage  microstorage.Storage
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
		}

		host, err = framework.NewHost(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	k8sExtClient, err := apiextensionsclient.NewForConfig(host.RestConfig())
	if err != nil {
		return Config{}, microerror.Mask(err)
	}

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.Config{
			K8sExtClient: k8sExtClient,
			Logger:       logger,
		}

		crdClient, err = k8scrdclient.New(c)
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

	var crdStorage microstorage.Storage
	{
		c := crdstorage.Config{
			CRDClient: crdClient,
			G8sClient: host.G8sClient(),
			K8sClient: host.K8sClient(),
			Logger:    logger,

			Name: "kvm-e2e",
			Namespace: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			},
		}

		s, err := crdstorage.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}

		err = s.Boot(context.Background())
		if err != nil {
			return Config{}, microerror.Mask(err)
		}

		crdStorage = s
	}

	var retryStorage microstorage.Storage
	{
		c := retrystorage.Config{
			Logger:     logger,
			Underlying: crdStorage,
		}

		retryStorage, err = retrystorage.New(c)
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
		Storage:  retryStorage,
	}

	return c, nil
}
