package readylabel

import (
	"reflect"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Name = "readylabel"
)

type Config struct {
	Cluster             v1alpha1.KVMConfig
	ManagementK8sClient k8sclient.Interface
	Logger              micrologger.Logger
}

type Resource struct {
	cluster             v1alpha1.KVMConfig
	managementK8sClient client.Client
	logger              micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if reflect.DeepEqual(config.Cluster, v1alpha1.KVMConfig{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.Cluster must not be empty", config)
	}
	if config.ManagementK8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ManagementK8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		cluster:             config.Cluster,
		managementK8sClient: config.ManagementK8sClient.CtrlClient(),
		logger:              config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
