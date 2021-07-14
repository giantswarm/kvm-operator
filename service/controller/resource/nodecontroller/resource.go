package nodecontroller

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	workloadcluster "github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/service/controller/internal/nodecontroller"
)

const (
	Name = "nodecontroller"
)

type Config struct {
	CtrlClient      client.Client
	Logger          micrologger.Logger
	WorkloadCluster workloadcluster.Interface
}

type Resource struct {
	ctrlClient      client.Client
	logger          micrologger.Logger
	workloadCluster workloadcluster.Interface

	controllerMutex sync.Mutex // Used to protect controllers map from concurrent access
	controllers     map[types.NamespacedName]*nodecontroller.Controller
}

func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.WorkloadCluster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.WorkloadCluster must not be empty", config)
	}

	r := &Resource{
		ctrlClient:      config.CtrlClient,
		logger:          config.Logger,
		workloadCluster: config.WorkloadCluster,

		controllers: map[types.NamespacedName]*nodecontroller.Controller{},
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		r.Stop()
	}()

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Stop() {
	r.controllerMutex.Lock()
	for _, controller := range r.controllers {
		controller.Stop()
	}
	r.controllerMutex.Unlock()
}

func controllerMapKey(cluster v1alpha1.KVMConfig) types.NamespacedName {
	return types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name,
	}
}
