package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

const (
	Name = "endpoint"
)

type Config struct {
	CtrlClient client.Client
	Logger     micrologger.Logger
}

type Resource struct {
	ctrlClient client.Client
	logger     micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) podEndpointData(ctx context.Context, pod corev1.Pod) (nodeIP string, serviceName types.NamespacedName, err error) {
	nodeIP = pod.Annotations[key.AnnotationIp]
	if nodeIP == "" {
		r.logger.Debugf(ctx, "node pod has no ip annotation %#q, skipping", key.AnnotationIp)
		err = microerror.Mask(missingAnnotationError)
		return
	}

	serviceName = types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Annotations[key.AnnotationService],
	}
	if serviceName.Name == "" {
		r.logger.Debugf(ctx, "node pod has no service annotation %#q, skipping", key.AnnotationService)
		err = microerror.Mask(missingAnnotationError)
		return
	}

	return
}
