package cleanupendpointips

import (
	"context"
	"fmt"
	"sort"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	// At first we need to create a Kubernetes client for the reconciled tenant
	// cluster.

	r.logger.LogCtx(ctx, "level", "debug", "message", "creating Kubernetes client for tenant cluster")

	k8sClient, err := key.CreateK8sClientForTenantCluster(ctx, obj, r.logger, r.tenantCluster)
	if tenantcluster.IsTimeout(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not create Kubernetes client for tenant cluster")
		r.logger.LogCtx(ctx, "level", "debug", "message", "waiting for certificates timed out")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	} else if tenant.IsAPINotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "tenant cluster is not available")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "created Kubernetes client for tenant cluster")

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
			epRemoved, masterEndpoint, err := removeDeadIPFromEndpoints(masterEndpoint, nodes, pods)
			if err != nil {
				return microerror.Mask(err)
			}
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
			epRemoved, workerEndpoint, err := removeDeadIPFromEndpoints(workerEndpoint, nodes, pods)
			if err != nil {
				return microerror.Mask(err)
			}
			if epRemoved > 0 {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("removing %d dead ips from the worker endpoints", epRemoved))

				// E 06/09 14:54:30 /apis/provider.giantswarm.io/v1alpha1/namespaces/default/kvmconfigs/73j3v failed processing event | operatorkit/controller/controller.go:513 | controller=kvm-operator-deleter | event=update | loop=170 | stack=map[annotation:
				// Endpoints "worker" is invalid: subsets[0]: Required value: must specify `addresses` or `notReadyAddresses` kind:unknown
				// stack:[map[file:/root/project/service/controller/resource/cleanupendpointips/create.go line:110]
				// map[file:/go/pkg/mod/github.com/giantswarm/operatorkit@v0.2.1/resource/wrapper/retryresource/basic_resource.go line:52]
				// map[file:/go/pkg/mod/github.com/giantswarm/backoff@v0.2.0/retry.go line:23]
				// map[file:/go/pkg/mod/github.com/giantswarm/operatorkit@v0.2.1/resource/wrapper/retryresource/basic_resource.go line:64]
				// map[file:/go/pkg/mod/github.com/giantswarm/operatorkit@v0.2.1/resource/wrapper/metricsresource/basic_resource.go line:43]
				// map[file:/go/pkg/mod/github.com/giantswarm/operatorkit@v0.2.1/controller/controller.go line:619]]]

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting workerEndpoint: %v#!", workerEndpoint))

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

func controlPlanePodForTCNode(node corev1.Node, pods []corev1.Pod) (corev1.Pod, error) {
	for _, pod := range pods {
		if pod.Name == node.Name {
			return pod, nil
		}
	}
	// Unless there is a race condition where the Pods are modified
	// after being checked by podsEqualNodes(), this should never be reached.
	return corev1.Pod{}, microerror.Maskf(noPodForNodeError, fmt.Sprintf("no control plane pod for tenant cluster node %s", node.Name))
}

// removeDeadIPFromEndpoints compares endpoint IPs with current state of nodes and
// removes any IP addresses that does not belong to any node.
func removeDeadIPFromEndpoints(endpoints *corev1.Endpoints, nodes []corev1.Node, cpPods []corev1.Pod) (int, *corev1.Endpoints, error) {
	endpointAddresses := endpoints.Subsets[0].Addresses

	var indexesToDelete []int
	for i, ip := range endpointAddresses {
		found := false
		// check if the ip belongs to any k8s node
		for _, node := range nodes {

			if node.Labels["ip"] == ip.IP {
				// Find the control plane pod representing this node
				cpPod, err := controlPlanePodForTCNode(node, cpPods)
				if err != nil {
					return len(indexesToDelete), endpoints, microerror.Mask(err)
				}

				// Check if the CP pod is Ready
				for _, c := range cpPod.Status.Conditions {
					if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
						// Keep this Pod in our endpoints
						found = true
						break
					}
				}

				// Otherwise, let this pod be removed
			}
		}
		// endpoint ip does not belong to any node with a healthy CP pod, lets remove it
		if !found {
			indexesToDelete = append(indexesToDelete, i)
		}
	}
	if len(indexesToDelete) > 0 {
		endpoints.Subsets[0].Addresses = removeFromEndpointAddressList(endpointAddresses, indexesToDelete)
	}
	return len(indexesToDelete), endpoints, nil
}
