package node

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/api/v1/node"

	"github.com/giantswarm/kvm-operator/service/controller/v14/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// At first we need to create a Kubernetes client for the reconciled guest
	// cluster.
	var k8sClient kubernetes.Interface
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "creating K8s client for the guest cluster")

		certs, err := r.certsSearcher.SearchCluster(key.ClusterID(customObject))
		if err != nil {
			return microerror.Mask(err)
		}

		c := k8srestconfig.Config{
			Logger: r.logger,

			Address:   key.ClusterAPIEndpoint(customObject),
			InCluster: false,
			TLS: k8srestconfig.TLSClientConfig{
				CAData:  certs.APIServer.CA,
				CrtData: certs.APIServer.Crt,
				KeyData: certs.APIServer.Key,
			},
		}
		restConfig, err := k8srestconfig.New(c)
		if err != nil {
			return microerror.Mask(err)
		}

		k8sClient, err = kubernetes.NewForConfig(restConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "created K8s client for the guest cluster")
	}

	// We need to fetch the nodes being registered within the guest cluster's
	// Kubernetes API. The list of nodes is used below to sort out which ones have
	// to be deleted if there does no associated host cluster pod exist.
	var nodes []corev1.Node
	{
		list, err := k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
		if guest.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster is not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource for custom object")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
		nodes = list.Items
	}

	// Fetch the list of pods running on the host cluster. These pods serve VMs
	// which in turn run the guest cluster nodes. We use the pods to compare them
	// against the guest cluster nodes below.
	var pods []corev1.Pod
	{
		n := key.ClusterID(customObject)
		list, err := r.k8sClient.CoreV1().Pods(n).List(metav1.ListOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		pods = list.Items
	}

	// Iterate through all nodes and compare them against the pods of the host
	// cluster. Nodes being in a Ready state are fine. Nodes that belong to host
	// cluster pods are also ok. If a guest cluster node does not have an
	// associated host cluster pod, we delete it from the guest cluster's
	// Kubernetes API.
	for _, n := range nodes {
		if node.IsNodeReady(&n) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting node '%s' because it is in state 'Ready'", n.GetName()))
			continue
		}

		if doesNodeExistAsPod(pods, n) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting node '%s' because its host cluster pod does exist", n.GetName()))
			continue
		}

		if isPodOfNodeRunning(pods, n) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting node '%s' because its host cluster pod is running", n.GetName()))
			continue
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting node '%s' in the guest cluster's Kubernetes API", n.GetName()))

		err = k8sClient.CoreV1().Nodes().Delete(n.GetName(), nil)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted node '%s' in the guest cluster's Kubernetes API", n.GetName()))
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
