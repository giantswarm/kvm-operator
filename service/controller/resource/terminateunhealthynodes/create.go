package terminateunhealthynodes

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/annotation"
	"github.com/giantswarm/badnodedetector/pkg/detector"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

const (
	nodeTerminationTickThreshold = 6
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// check for annotation enabling the node auto repair feature
	if _, ok := customResource.Annotations[annotation.NodeTerminateUnhealthy]; !ok {
		r.logger.Debugf(ctx, "terminate unhealthy node feature is not enabled for this cluster")
		return nil
	}

	var tcCtrlClient client.Client
	{
		tcCtrlClient, err = key.CreateCtrlClientForTenantCluster(ctx, obj, r.logger, r.tenantCluster)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var detectorService *detector.Detector
	{
		detectorConfig := detector.Config{
			K8sClient: tcCtrlClient,
			Logger:    r.logger,

			NotReadyTickThreshold: nodeTerminationTickThreshold,
		}

		detectorService, err = detector.NewDetector(detectorConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	nodesToTerminate, err := detectorService.DetectBadNodes(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(nodesToTerminate) > 0 {
		for _, n := range nodesToTerminate {
			err := r.terminateNode(ctx, customResource.Spec.Cluster.ID, n)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		// reset tick counters on all nodes in cluster to have a graceful period after terminating nodes
		err := detectorService.ResetTickCounters(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.Debugf(ctx, "resetting tick node counters on all nodes in tenant cluster")
	}

	return nil
}

func (r *Resource) terminateNode(ctx context.Context, clusterID string, node corev1.Node) error {
	r.logger.Debugf(ctx, "getting corresponding CP pod for node %s", node.Name)
	pod, err := r.k8sClient.CoreV1().Pods(clusterID).Get(ctx, node.Name, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "terminating unhealhty node %s", node.Name)
	err = r.k8sClient.CoreV1().Pods(clusterID).Delete(ctx, pod.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	r.logger.Debugf(ctx, "terminated unhealhty node %s", node.Name)

	return nil
}
