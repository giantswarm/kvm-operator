package configmap

import (
	"reflect"

	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys/v2"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/service/controller/cloudconfig"
)

const (
	KeyUserData = "user_data"
	// Name is the identifier of the resource.
	Name = "configmap"
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

func containsConfigMap(list []*corev1.ConfigMap, item *corev1.ConfigMap) bool {
	_, err := getConfigMapByName(list, item.Name)
	return err == nil
}

// equals asseses the equality of ConfigMaps with regards to distinguishing
// fields.
func equals(a, b *corev1.ConfigMap) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Namespace != b.Namespace {
		return false
	}
	if !reflect.DeepEqual(a.Data, b.Data) {
		return false
	}
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return false
	}

	return true
}

func getConfigMapByName(list []*corev1.ConfigMap, name string) (*corev1.ConfigMap, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

// isEmpty checks if a ConfigMap is empty.
func isEmpty(c *corev1.ConfigMap) bool {
	if c == nil {
		return true
	}

	return equals(c, &corev1.ConfigMap{})
}

func toConfigMaps(v interface{}) ([]*corev1.ConfigMap, error) {
	if v == nil {
		return nil, nil
	}

	configMaps, ok := v.([]*corev1.ConfigMap)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*corev1.ConfigMap{}, v)
	}

	return configMaps, nil
}
