package node

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/api/v1/node"

	"github.com/giantswarm/kvm-operator/service/controller/v22/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// At first we need to create a Kubernetes client for the reconciled tenant
	// cluster.
	var k8sClient kubernetes.Interface
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating Kubernetes client for tenant cluster")

		i := key.ClusterID(customObject)
		e := key.ClusterAPIEndpoint(customObject)

		k8sClient, err = r.tenantCluster.NewK8sClient(ctx, i, e)
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not create Kubernetes client for tenant cluster")
			r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for certificates timed out")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "created Kubernetes client for tenant cluster")
		}
	}

	// We need to fetch the nodes being registered within the tenant cluster's
	// Kubernetes API. The list of nodes is used below to sort out which ones have
	// to be deleted if there does no associated control plane pod exist.
	var nodes []corev1.Node
	{
		list, err := k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
		nodes = list.Items
	}

	// Fetch the list of pods running on the control plane. These pods serve VMs
	// which in turn run the tenant cluster nodes. We use the pods to compare them
	// against the tenant cluster nodes below.
	var pods []corev1.Pod
	{
		n := key.ClusterID(customObject)
		list, err := r.k8sClient.CoreV1().Pods(n).List(metav1.ListOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		pods = list.Items
	}

	// Iterate through all nodes and compare them against the pods of the control
	// plane. Nodes being in a Ready state are fine. Nodes that belong to control
	// plane pods are also ok. If a tenant cluster node does not have an
	// associated control plane pod, we delete it from the tenant cluster's
	// Kubernetes API.
	for _, n := range nodes {
		if node.IsNodeReady(&n) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting node '%s' because it is in state 'Ready'", n.GetName()))
			continue
		}

		if doesNodeExistAsPod(pods, n) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting node '%s' because its control plane pod does exist", n.GetName()))
			continue
		}

		if isPodOfNodeRunning(pods, n) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting node '%s' because its control plane pod is running", n.GetName()))
			continue
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting node '%s' in the tenant cluster's Kubernetes API", n.GetName()))

		err = k8sClient.CoreV1().Nodes().Delete(n.GetName(), nil)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted node '%s' in the tenant cluster's Kubernetes API", n.GetName()))
	}

	return nil
}

func doesNodeExistAsPod(pods []corev1.Pod, n corev1.Node) bool {
	for _, p := range pods {
		if p.GetName() == n.GetName() {
			return true
		}
	}

	return false
}

func isPodOfNodeRunning(pods []corev1.Pod, n corev1.Node) bool {
	for _, p := range pods {
		if p.GetName() == n.GetName() {
			if p.Status.Phase == corev1.PodRunning {
				return true
			}
		}
	}

	return false
}
