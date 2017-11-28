package configmap

import (
	"reflect"

	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
<<<<<<< HEAD
	"github.com/giantswarm/randomkeytpr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
=======
>>>>>>> origin/master
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/kvm-operator/service/cloudconfig"
)

const (
	KeyUserData = "user_data"
	// Name is the identifier of the resource.
	Name = "configmap"
)

// Config represents the configuration used to create a new config map resource.
type Config struct {
	// Dependencies.
	CertWatcher certificatetpr.Searcher
	CloudConfig *cloudconfig.CloudConfig
	K8sClient   kubernetes.Interface
	KeyWatcher  randomkeytpr.Searcher
	Logger      micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new config map
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		CertWatcher: nil,
		CloudConfig: nil,
		K8sClient:   nil,
		KeyWatcher:  nil,
		Logger:      nil,
	}
}

// Resource implements the config map resource.
type Resource struct {
	// Dependencies.
	certWatcher certificatetpr.Searcher
	cloudConfig *cloudconfig.CloudConfig
	k8sClient   kubernetes.Interface
	keyWatcher  randomkeytpr.Searcher
	logger      micrologger.Logger
}

// New creates a new configured config map resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.CertWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertWatcher must not be empty")
	}
	if config.CloudConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CloudConfig must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.KeyWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.KeyWatcher must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		certWatcher: config.CertWatcher,
		cloudConfig: config.CloudConfig,
		k8sClient:   config.K8sClient,
		keyWatcher:  config.KeyWatcher,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func containsConfigMap(list []*apiv1.ConfigMap, item *apiv1.ConfigMap) bool {
	_, err := getConfigMapByName(list, item.Name)
	if err != nil {
		return false
	}

	return true
}

func getConfigMapByName(list []*apiv1.ConfigMap, name string) (*apiv1.ConfigMap, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, microerror.Mask(notFoundError)
}

func isConfigMapModified(a, b *apiv1.ConfigMap) bool {
	return !reflect.DeepEqual(a.Data, b.Data)
}

func toConfigMaps(v interface{}) ([]*apiv1.ConfigMap, error) {
	if v == nil {
		return nil, nil
	}

	configMaps, ok := v.([]*apiv1.ConfigMap)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []*apiv1.ConfigMap{}, v)
	}

	return configMaps, nil
}
