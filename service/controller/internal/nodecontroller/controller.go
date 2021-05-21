package nodecontroller

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/micrologger/loggermeta"
	operatorkitcontroller "github.com/giantswarm/operatorkit/v4/pkg/controller"
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

const (
	ResyncPeriod = operatorkitcontroller.DefaultResyncPeriod

	loggerKeyController = "controller"
	loggerKeyEvent      = "event"
	loggerKeyObject     = "object"
	loggerKeyVersion    = "version"
)

type Config struct {
	Cluster             v1alpha1.KVMConfig
	ManagementK8sClient client.Client
	WorkloadK8sClient   k8sclient.Interface
	Logger              micrologger.Logger

	Name     string
	Selector labels.Selector
}

// This controller is based on the operatorkit controller but cuts out metrics collectors, Sentry, and other undesirable
// features for a dynamic workload cluster controller.
type Controller struct {
	managementK8sClient client.Client
	workloadK8sClient   k8sclient.Interface
	logger              micrologger.Logger

	stopOnce       sync.Once
	stopped        chan struct{}
	lastReconciled time.Time

	name     string
	selector labels.Selector
	cluster  v1alpha1.KVMConfig
}

// New creates a new configured workload cluster node controller.
func New(config Config) (*Controller, error) {
	if reflect.DeepEqual(config.Cluster, v1alpha1.KVMConfig{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.Cluster must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ManagementK8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ManagementK8sClient must not be empty", config)
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Name must not be empty", config)
	}
	if config.Selector == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Selector must not be empty", config)
	}
	if config.WorkloadK8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.WorkloadK8sClient must not be empty", config)
	}

	c := &Controller{
		managementK8sClient: config.ManagementK8sClient,
		workloadK8sClient:   config.WorkloadK8sClient,
		logger:              config.Logger,

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
			MetricsBindAddress: operatorkitcontroller.DisableMetricsServing,
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
				DeleteFunc:  func(e event.DeleteEvent) bool { return c.selector.Matches(labels.Set(e.Meta.GetLabels())) },
				UpdateFunc:  func(e event.UpdateEvent) bool { return c.selector.Matches(labels.Set(e.MetaNew.GetLabels())) },
				GenericFunc: func(e event.GenericEvent) bool { return c.selector.Matches(labels.Set(e.Meta.GetLabels())) },
			}).
			Complete(c)
		if err != nil {
			return microerror.Mask(err)
		}

		// Start in goroutine so we don't block caller
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

// Equal returns true when the given controllers are equal as it relates to watching the workload
// cluster Kubernetes APIs.
func (c *Controller) Equal(other *Controller) bool {
	// avoid nil pointer dereference panic
	if c != nil && other == nil {
		return false
	}

	// compare rest configs
	// rest.Config equality based on setup function here https://github.com/giantswarm/tenantcluster/blob/3531fb3d3698c0a69ab51f42c95207cb80761529/pkg/tenantcluster/tenantcluster.go#L72-L82
	thisRestConfig := c.workloadK8sClient.RESTConfig()
	otherRestConfig := other.workloadK8sClient.RESTConfig()
	if thisRestConfig.Host != otherRestConfig.Host {
		return false
	}
	if !reflect.DeepEqual(thisRestConfig.TLSClientConfig, otherRestConfig.TLSClientConfig) {
		return false
	}

	// compare kvmconfigs
	if !reflect.DeepEqual(c.cluster.Spec, other.cluster.Spec) {
		return false
	}

	// compare selectors
	if c.selector.String() != other.selector.String() {
		return false
	}

	// compare names
	if c.name != other.name {
		return false
	}

	return true
}

func (c *Controller) LastReconciled() time.Time {
	return c.lastReconciled
}

func setLoggerCtxValue(ctx context.Context, key, value string) context.Context {
	m, ok := loggermeta.FromContext(ctx)
	if !ok {
		m = loggermeta.New()
		ctx = loggermeta.NewContext(ctx, m)
	}

	m.KeyVals[key] = value

	return ctx
}

func (c *Controller) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	ctx = setLoggerCtxValue(ctx, loggerKeyController, c.name)

	var node corev1.Node
	err := c.workloadK8sClient.CtrlClient().Get(ctx, req.NamespacedName, &node)
	if errors.IsNotFound(err) {
		return key.RequeueNone, nil
	} else if err != nil {
		return key.RequeueErrorLong, microerror.Mask(err)
	}

	ctx = setLoggerCtxValue(ctx, loggerKeyObject, node.GetSelfLink())
	ctx = setLoggerCtxValue(ctx, loggerKeyVersion, node.GetResourceVersion())

	if key.IsDeleted(&node) {
		ctx = setLoggerCtxValue(ctx, loggerKeyEvent, "delete")
		result, err := c.ensureDeleted(ctx, node)
		if err != nil {
			return result, microerror.Mask(err)
		}
	} else {
		ctx = setLoggerCtxValue(ctx, loggerKeyEvent, "create")
		result, err := c.ensureCreated(ctx, node)
		if err != nil {
			return result, microerror.Mask(err)
		}
	}

	c.lastReconciled = time.Now()

	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: ResyncPeriod,
	}, nil
}

func (c *Controller) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopped)
	})
}
