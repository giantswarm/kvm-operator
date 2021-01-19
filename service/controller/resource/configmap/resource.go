package configmap

import (
	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/service/controller/cloudconfig"
)

const (
	// Name is the identifier of the resource.
	Name = "configmap"

	KeyUserData = "user_data"
)

// Config represents the configuration used to create a new config map resource.
type Config struct {
	// Dependencies.
	CertsSearcher  certs.Interface
	CloudConfig    *cloudconfig.CloudConfig
	CtrlClient     client.Client
	KeyWatcher     randomkeys.Interface
	Logger         micrologger.Logger
	DockerhubToken string
	RegistryDomain string
}

// Resource implements the config map resource.
type Resource struct {
	// Dependencies.
	certsSearcher  certs.Interface
	cloudConfig    *cloudconfig.CloudConfig
	ctrlClient     client.Client
	keyWatcher     randomkeys.Interface
	logger         micrologger.Logger
	registryDomain string
}

// New creates a new configured config map resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CertsSearcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertsSearcher must not be empty")
	}
	if config.CloudConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CloudConfig must not be empty")
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CtrlClient must not be empty")
	}
	if config.KeyWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.KeyWatcher must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}
	if config.DockerhubToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DockerhubToken must not be empty", config)
	}

	newService := &Resource{
		// Dependencies.
		certsSearcher:  config.CertsSearcher,
		cloudConfig:    config.CloudConfig,
		ctrlClient:     config.CtrlClient,
		keyWatcher:     config.KeyWatcher,
		logger:         config.Logger,
		registryDomain: config.RegistryDomain,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}
