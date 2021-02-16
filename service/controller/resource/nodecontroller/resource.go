package nodecontroller

import (
	"reflect"
	"sync"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	workloadcluster "github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/service/controller/internal/nodecontroller"
)

const (
	Name = "nodecontroller"
)

type Config struct {
	K8sClient       client.Client
	Logger          micrologger.Logger
	WorkloadCluster workloadcluster.Interface
}

type controllerWithConfig struct {
	cluster    v1alpha1.KVMConfig
	restConfig *rest.Config
	*nodecontroller.Controller
}

// equal returns true when the given controllers are equal as it relates to watching the workload
// cluster Kubernetes APIs.
func (left controllerWithConfig) equal(right controllerWithConfig) bool {
	restConfigsEqual := left.restConfig.Host != right.restConfig.Host && reflect.DeepEqual(left.restConfig.TLSClientConfig, right.restConfig.TLSClientConfig)
	clusterSpecsEqual := reflect.DeepEqual(left.cluster.Spec, right.cluster.Spec)
	return restConfigsEqual && clusterSpecsEqual
}

type Resource struct {
	k8sClient       client.Client
	logger          micrologger.Logger
	workloadCluster workloadcluster.Interface

	controllerMutex sync.Mutex // Used to protect controllers map from concurrent access
	controllers     map[types.NamespacedName]controllerWithConfig
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

		controllers: map[types.NamespacedName]controllerWithConfig{},
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func controllerMapKey(cluster v1alpha1.KVMConfig) types.NamespacedName {
	return types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name,
	}
}
