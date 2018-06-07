package node

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/api/v1/node"
	"k8s.io/kubernetes/pkg/cloudprovider"

	"github.com/giantswarm/kvm-operator/service/controller/v13/key"
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

	// Fetch the list of instances from the cloud provider. This is a list of host
	// cluster nodes which we use to compare against the guest cluster nodes
	// below.
	var instances cloudprovider.Instances
	{
		var ok bool

		instances, ok = r.cloudProvider.Instances()
		if !ok {
			return microerror.Mask(instanceNotFoundError)
		}
	}

	// We need to fetch the nodes being registered within the guest cluster's
	// Kubernetes API. The list of nodes is used below to sort out which ones have
	// to be deleted if there does no associated host cluster node exist.
	var nodes []corev1.Node
	{
		list, err := k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		nodes = list.Items
	}

	// Iterate through all nodes and compare them against the instances of the
	// cloud provider API. Nodes being in a Ready state are fine. Nodes that
	// belong to host cluster nodes are also ok. If a guest cluster node does not
	// have an associated host cluster node we delete it from the guest cluster's
	// Kubernetes API.
	for _, n := range nodes {
		if node.IsNodeReady(&n) {
			continue
		}

		exists, err := doesNodeExistAsInstance(instances, n)
		if err != nil {
			return microerror.Mask(err)
		}
		if exists {
			continue
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleting the node in the guest cluster's Kubernetes API")

		err = k8sClient.CoreV1().Nodes().Delete(n.GetName(), nil)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted the node in the guest cluster's Kubernetes API")
	}

	return nil
}

func doesNodeExistAsInstance(instances cloudprovider.Instances, n corev1.Node) (bool, error) {
	_, err := instances.ExternalID(types.NodeName(n.Name))
	if IsInstanceNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}
