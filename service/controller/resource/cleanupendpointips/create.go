package cleanupendpointips

import (
	"context"
	"fmt"
	"sort"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/kvm-operator/service/controller/key"
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

		restConfig, err := r.tenantCluster.NewRestConfig(ctx, i, e)
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not create Kubernetes client for tenant cluster")
			r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for certificates timed out")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
		clientsConfig := k8sclient.ClientsConfig{
			Logger:     r.logger,
			RestConfig: restConfig,
		}
		k8sClients, err := k8sclient.NewClients(clientsConfig)
		if tenant.IsAPINotAvailable(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		k8sClient = k8sClients.K8sClient()
		r.logger.LogCtx(ctx, "level", "debug", "message", "created Kubernetes client for tenant cluster")
	}

	// We need to fetch the nodes being registered within the tenant cluster's
	// Kubernetes API.
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
	// which in turn run the tenant cluster nodes.
	var pods []corev1.Pod
	{
		n := key.ClusterID(customObject)
		list, err := r.k8sClient.CoreV1().Pods(n).List(metav1.ListOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		pods = list.Items
	}
	// Check if all k8s-kvm pods in CP are registered as nodes in the TC.
	if podsEqualNodes(pods, nodes) {
		n := key.ClusterID(customObject)

		{
			masterEndpoint, err := r.k8sClient.CoreV1().Endpoints(n).Get(key.MasterID, metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			epRemoved, masterEndpoint := removeDeadIPFromEndpoints(masterEndpoint, nodes)
			if epRemoved > 0 {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("removing %d dead ips from the master endpoints", epRemoved))

				_, err = r.k8sClient.CoreV1().Endpoints(n).Update(masterEndpoint)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}

		{
			workerEndpoint, err := r.k8sClient.CoreV1().Endpoints(n).Get(key.WorkerID, metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			epRemoved, workerEndpoint := removeDeadIPFromEndpoints(workerEndpoint, nodes)
			if epRemoved > 0 {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("removing %d dead ips from the worker endpoints", epRemoved))

				_, err = r.k8sClient.CoreV1().Endpoints(n).Update(workerEndpoint)
				if err != nil {
					return microerror.Mask(err)
				}
			}
		}
	}
	return nil
}

// podsEqualNodes check if all k8s-kvm pods in CP are registered as nodes in the TC.
func podsEqualNodes(pods []corev1.Pod, nodes []corev1.Node) bool {
	if len(pods) != len(nodes) {
		return false
	}

	// sort pods and nodes by name
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].Name < pods[j].Name
	})
	sort.Slice(nodes, func(i, j int) bool {
		return pods[i].Name < pods[j].Name
	})

	for i := 0; i < len(pods); i++ {
		if pods[i].Name != nodes[i].Name {
			return false
		}
	}
	return true
}

// removeFromEndpointAddressList is removing slice elements from `addresses` defined by `indexesToRemove`.
func removeFromEndpointAddressList(addresses []corev1.EndpointAddress, indexesToRemove []int) []corev1.EndpointAddress {
	var newAddresses []corev1.EndpointAddress
	for i, ip := range addresses {
		remove := false
		for _, j := range indexesToRemove {
			if i == j {
				remove = true
			}
		}
		if !remove {
			newAddresses = append(newAddresses, ip)
		}
	}
	return newAddresses
}

// removeDeadIPFromEndpoints compares endpoint IPs with current state of nodes and
// removes any IP addresses that does not belong to any node.
func removeDeadIPFromEndpoints(endpoints *corev1.Endpoints, nodes []corev1.Node) (int, *corev1.Endpoints) {
	endpointAddresses := endpoints.Subsets[0].Addresses

	var indexesToDelete []int
	for i, ip := range endpointAddresses {
		found := false
		// check if the ip belongs to any k8s node
		for _, node := range nodes {
			if node.Labels["ip"] == ip.IP {
				found = true
				break
			}
		}
		// endpoint ip does not belong to any node, lets remove it
		if !found {
			indexesToDelete = append(indexesToDelete, i)
		}
	}
	if len(indexesToDelete) > 0 {
		endpoints.Subsets[0].Addresses = removeFromEndpointAddressList(endpointAddresses, indexesToDelete)
	}
	return len(indexesToDelete), endpoints
}
