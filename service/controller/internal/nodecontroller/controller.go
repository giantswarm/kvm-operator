package nodecontroller

import (
	"context"
	"reflect"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/to"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

type Config struct {
	Cluster             v1alpha1.KVMConfig
	ManagementK8sClient client.Client
	WorkloadK8sClient   k8sclient.Interface
	Logger              micrologger.Logger

	Name     string
	Selector labels.Selector
}

const (
	ResyncPeriod = 10 * time.Minute
)

const (
	DisableMetricsServing = "0"
)

type Controller struct {
	managementK8sClient client.Client
	workloadK8sClient   k8sclient.Interface
	logger              micrologger.Logger

	booted         chan struct{}
	stopped        chan struct{}
	lastReconciled time.Time

	name     string
	selector labels.Selector
	cluster  v1alpha1.KVMConfig
}

// New creates a new configured operator controller.
func New(config Config) (*Controller, error) {
	if config.ManagementK8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ManagementK8sClient must not be empty", config)
	}
	if config.WorkloadK8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.WorkloadK8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Name must not be empty", config)
	}

	c := &Controller{
		managementK8sClient: config.ManagementK8sClient,
		workloadK8sClient:   config.WorkloadK8sClient,
		logger:              config.Logger,

		booted:         make(chan struct{}),
		stopped:        make(chan struct{}),
		lastReconciled: time.Time{},
		name:           config.Name,
		selector:       config.Selector,
		cluster:        config.Cluster,
	}

	return c, nil
}

func (c *Controller) Boot() error {
	var mgr manager.Manager

	{
		o := manager.Options{
			// MetricsBindAddress is set to 0 in order to disable it. We do this
			// ourselves.
			MetricsBindAddress: DisableMetricsServing,
			SyncPeriod:         to.DurationP(ResyncPeriod),
		}

		var err error
		mgr, err = manager.New(c.workloadK8sClient.RESTConfig(), o)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		// We build our controller and set up its reconciliation.
		// We use the Complete() method instead of Build() because we don't
		// need the controller instance.
		err := builder.
			ControllerManagedBy(mgr).
			For(&corev1.Node{}).
			WithOptions(controller.Options{
				MaxConcurrentReconciles: 1,
				Reconciler:              c,
			}).
			WithEventFilter(predicate.Funcs{
				CreateFunc:  func(e event.CreateEvent) bool { return c.selector.Matches(labels.Set(e.Meta.GetLabels())) },
				DeleteFunc:  func(e event.DeleteEvent) bool { return false },
				UpdateFunc:  func(e event.UpdateEvent) bool { return c.selector.Matches(labels.Set(e.MetaNew.GetLabels())) },
				GenericFunc: func(e event.GenericEvent) bool { return c.selector.Matches(labels.Set(e.Meta.GetLabels())) },
			}).
			Complete(c)
		if err != nil {
			return microerror.Mask(err)
		}

		// We put the controller into a booted state by closing its booted
		// channel once so users know when to go ahead.
		select {
		case <-c.booted:
		default:
			close(c.booted)
		}

		go func() {
			// mgr.Start() blocks the boot process until it ends gracefully or fails.
			err = mgr.Start(c.stopped)
			if err != nil {
				panic(microerror.JSON(err))
			}
		}()
	}

	return nil
}

func (c *Controller) Booted() chan struct{} {
	return c.booted
}

// Equal returns true when the given controllers are equal as it relates to watching the workload
// cluster Kubernetes APIs.
func (c *Controller) Equal(other *Controller) bool {
	thisRestConfig := c.workloadK8sClient.RESTConfig()
	otherRestConfig := other.workloadK8sClient.RESTConfig()
	if thisRestConfig.Host != otherRestConfig.Host {
		return false
	}
	if reflect.DeepEqual(thisRestConfig.TLSClientConfig, otherRestConfig.TLSClientConfig) {
		return false
	}
	if !reflect.DeepEqual(c.cluster.Spec, other.cluster.Spec) {
		return false
	}
	return true
}

func (c *Controller) LastReconciled() time.Time {
	return c.lastReconciled
}

func (c *Controller) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()

	var node corev1.Node
	err := c.workloadK8sClient.CtrlClient().Get(ctx, req.NamespacedName, &node)
	if errors.IsNotFound(err) {
		return reconcile.Result{Requeue: false}, nil
	} else if err != nil {
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 30,
		}, microerror.Mask(err)
	}

	if key.IsDeleted(&node) {
		return reconcile.Result{Requeue: false}, nil
	}

	result, err := c.ensureConditions(ctx, node)
	if err != nil {
		return result, microerror.Mask(err)
	}

	c.lastReconciled = time.Now()

	return reconcile.Result{Requeue: false}, nil
}

func (c *Controller) Stop() {
	close(c.stopped)
}
