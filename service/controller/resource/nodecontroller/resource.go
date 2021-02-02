package nodecontroller

import (
	"sync"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	workloadcluster "github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/kvm-operator/service/controller/internal/nodecontroller"
)

const (
	Name = "nodecontroller"
)

type Config struct {
	K8sClient       k8sclient.Interface
	Logger          micrologger.Logger
	WorkloadCluster workloadcluster.Interface
}

type Resource struct {
	k8sClient       k8sclient.Interface
	logger          micrologger.Logger
	workloadCluster workloadcluster.Interface
	controllerMutex sync.Mutex
	controllers     map[types.NamespacedName]*nodecontroller.Controller
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.WorkloadCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.WorkloadCluster must not be empty", config)
	}

	r := &Resource{
		k8sClient:       config.K8sClient,
		logger:          config.Logger,
		workloadCluster: config.WorkloadCluster,
		controllers:     map[types.NamespacedName]*nodecontroller.Controller{},
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func nodeControllerKey(cluster v1alpha1.KVMConfig) types.NamespacedName {
	return types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name,
	}
}
