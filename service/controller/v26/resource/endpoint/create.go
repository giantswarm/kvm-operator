package endpoint

import (
	"context"
	"fmt"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/kvm-operator/service/controller/v26/key"
	"github.com/giantswarm/tenantcluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createState interface{}) error {
	endpointToCreate, err := toK8sEndpoint(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	if !isEmptyEndpoint(endpointToCreate) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating endpoint '%s'", endpointToCreate.GetName()))

		_, err = r.k8sClient.CoreV1().Endpoints(endpointToCreate.Namespace).Create(endpointToCreate)
		if errors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created endpoint '%s'", endpointToCreate.GetName()))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not creating endpoint '%s'", endpointToCreate.GetName()))
	}

	return nil
}

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
	// Clear dead endpoint IPs only when the cluster is not in transitioning state.
	// The amount of nodes should be equal to amount of pods.
	if len(nodes) == len(pods) {
		n := key.ClusterID(customObject)

		{
			masterEndpoint, err := r.k8sClient.CoreV1().Endpoints(n).Get(key.MasterID, metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			epRemoved, masterEndpoint := r.removeDeadIPFromEndpoints(masterEndpoint, nodes)
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

			epRemoved, workerEndpoint := r.removeDeadIPFromEndpoints(workerEndpoint, nodes)
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

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var createChange *corev1.Endpoints
	{
		ips := ipsForCreateChange(currentEndpoint.IPs, desiredEndpoint.IPs)

		e := &Endpoint{
			Addresses:        ipsToAddresses(ips),
			IPs:              ips,
			Ports:            currentEndpoint.Ports,
			ResourceVersion:  currentEndpoint.ResourceVersion,
			ServiceName:      currentEndpoint.ServiceName,
			ServiceNamespace: currentEndpoint.ServiceNamespace,
		}

		createChange = r.newK8sEndpoint(e)
	}

	return createChange, nil
}

func ipsForCreateChange(currentIPs []string, desiredIPs []string) []string {
	var ips []string

	for _, ip := range desiredIPs {
		if !containsIP(ips, ip) {
			ips = append(ips, ip)
		}
	}

	if len(currentIPs) == 0 {
		return ips
	}

	return nil
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
func (r *Resource) removeDeadIPFromEndpoints(endpoints *corev1.Endpoints, nodes []corev1.Node) (int, *corev1.Endpoints) {
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
