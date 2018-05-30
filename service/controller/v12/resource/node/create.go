package node

import (
	"context"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/kvm-operator/service/controller/v11/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
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

	//	nodes, err := c.kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})

	//	_, currentReadyCondition = nodeutilv1.GetNodeCondition(&node.Status, v1.NodeReady)

	//	if currentReadyCondition.Status != v1.ConditionTrue {

	exists, err := doesNodeExistAsInstance(instances, node)
	if err != nil {
		return microerror.Mask(err)
	}

	// delete
	if !exists {
		r.logger.LogCtx(ctx, "debug", "deleting the node in the guest cluster's Kubernetes API")

		clusterID := key.ClusterID(customObject)
		apiEndpoint, err := key.ClusterAPIEndpoint(customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		k8sClient, err := r.newK8sClient(ctx, clusterID, apiDomain)
		if err != nil {
			return microerror.Mask(err)
		}

		err = k8sClient.CoreV1().Nodes().Delete(node.GetName(), nil)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "debug", "deleted the node in the guest cluster's Kubernetes API")
	}

	return nil
}

func (r *Resource) newK8sClient(ctx context.Context, clusterID, apiDomain string) (kubernetes.Interface, error) {
	r.logger.LogCtx(ctx, "level", "debug", "message", "creating K8s client for the guest cluster")

	restConfig, err := r.newRestConfig(ctx, clusterID, apiDomain)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "created K8s client for the guest cluster")

	return k8sClient, nil
}

func (r *Resource) newRestConfig(ctx context.Context, clusterID, apiDomain string) (*rest.Config, error) {
	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for certificate to connect to the guest cluster")

	operatorCerts, err := r.certsSearcher.SearchDraining(clusterID)
	if certs.IsTimeout(err) {
		return nil, microerror.Maskf(notFoundError, "cluster-operator cert not found for cluster")
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found certificate for connecting to the guest cluster")

	c := k8srestconfig.Config{
		Logger: r.logger,

		Address:   apiDomain,
		InCluster: false,
		TLS: k8srestconfig.TLSClientConfig{
			CAData:  operatorCerts.APIServer.CA,
			CrtData: operatorCerts.APIServer.Crt,
			KeyData: operatorCerts.APIServer.Key,
		},
	}
	restConfig, err := k8srestconfig.New(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return restConfig, nil
}

func doesNodeExistAsInstance(instances cloudprovider.Instances, node *v1.Node) (bool, error) {
	_, err := instances.ExternalID(types.NodeName(node.Name))
	if IsInstanceNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}
