package flannel

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/reconciliationcanceledcontext"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	namespace := key.FlannelNetworkNamespace(customObject)
	flannelNetwork, err := r.k8sClient.AppsV1().DaemonSets(namespace).Get(ctx, key.FlannelNetworkDaemonSetName, v1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	if flannelNetwork.Status.DesiredNumberScheduled != flannelNetwork.Status.NumberReady {
		r.logger.LogCtx(ctx, "level", "debug", "message", "flannel network is not ready, cancelling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
