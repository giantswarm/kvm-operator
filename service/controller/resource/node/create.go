package node

import (
	"context"

	workloaderrors "github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	workloadcluster "github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// At first we need to create a Kubernetes client for the reconciled workload
	// cluster.
	var ctrlClient client.Client
	{
		r.logger.Debugf(ctx, "creating Kubernetes client for workload cluster")

		k8sClients, err := key.CreateK8sClientForWorkloadCluster(ctx, customObject, r.logger, r.workloadCluster)
		if workloadcluster.IsTimeout(err) {
			r.logger.Debugf(ctx, "waiting for certificates timed out")
			return nil
		} else if workloaderrors.IsAPINotAvailable(err) || k8sclient.IsTimeout(err) {
			r.logger.Debugf(ctx, "workload cluster is not available")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		ctrlClient = k8sClients.CtrlClient()
		r.logger.Debugf(ctx, "created Kubernetes client for workload cluster")
	}

	// We need to fetch the nodes being registered within the workload cluster's
	// Kubernetes API. The list of nodes is used below to sort out which ones have
	// to be deleted if no associated management cluster pod exists.
	var nodes []corev1.Node
	{
		var list corev1.NodeList
		err := ctrlClient.List(ctx, &list)
		if workloaderrors.IsAPINotAvailable(err) {
			r.logger.Debugf(ctx, "workload cluster is not available")
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
		nodes = list.Items
	}

	// Fetch the list of pods running on the management cluster. These pods serve VMs
	// which in turn run the workload cluster nodes. We use the pods to compare them
	// against the workload cluster nodes below.
	var pods []corev1.Pod
	{
		var list corev1.PodList
		err := r.ctrlClient.List(ctx, &list, &client.ListOptions{
			Namespace: key.ClusterID(customObject),
		})
		if err != nil {
			return microerror.Mask(err)
		}
		pods = list.Items
	}

	// Iterate through all nodes and compare them against the pods of the management
	// cluster. Nodes being in a Ready state are fine. Nodes that belong to management
	// cluster pods are also ok. If a workload cluster node does not have an
	// associated management cluster pod, we delete it from the workload cluster's
	// Kubernetes API.
	for _, n := range nodes {
		if key.NodeIsReady(n) {
			r.logger.Debugf(ctx, "not deleting node '%s' because it is in state 'Ready'", n.GetName())
			continue
		}

		if doesNodeExistAsPod(pods, n) {
			r.logger.Debugf(ctx, "not deleting node '%s' because its management cluster pod does exist", n.GetName())
			continue
		}

		if isPodOfNodeRunning(pods, n) {
			r.logger.Debugf(ctx, "not deleting node '%s' because its management cluster pod is running", n.GetName())
			continue
		}

		r.logger.Debugf(ctx, "deleting node '%s' in the workload cluster's Kubernetes API", n.GetName())

		err = ctrlClient.Delete(ctx, &n) //nolint:gosec
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted node '%s' in the workload cluster's Kubernetes API", n.GetName())
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
