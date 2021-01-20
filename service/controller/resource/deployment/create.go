package deployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToKVMMachine(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := cr.Namespace
	var kvmCluster v1alpha2.KVMCluster
	{
		err := r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: clusterID,
			Name:      clusterID,
		}, &kvmCluster)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var existing v1.Deployment
	err = r.ctrlClient.Get(ctx, client.ObjectKey{
		Namespace: clusterID,
		Name:      clusterID,
	}, &existing)
	if apierrors.IsNotFound(err) {
		toCreate, err := r.newDeployment(ctx, kvmCluster, cr)
		if err != nil {
			return microerror.Mask(err)
		}
		err = r.ctrlClient.Create(ctx, toCreate)
		if err != nil {
			return microerror.Mask(err)
		}
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	desired, err := r.newDeployment(ctx, kvmCluster, cr)
	if err != nil {
		return microerror.Mask(err)
	}
	needsUpdate, err := r.needsUpdate(ctx, kvmCluster, &existing, desired)
	if err != nil {
		return microerror.Mask(err)
	}
	if needsUpdate {
		err = r.ctrlClient.Delete(ctx, &existing)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Resource) newDeployment(ctx context.Context, cluster v1alpha2.KVMCluster, machine v1alpha2.KVMMachine) (*v1.Deployment, error) {
	var release releasev1alpha1.Release
	{
		releaseVersion := machine.Labels[label.ReleaseVersion]
		var release releasev1alpha1.Release
		err := r.ctrlClient.Get(ctx, client.ObjectKey{
			Name: fmt.Sprintf("v%s", releaseVersion),
		}, &release)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var deployment *v1.Deployment
	role := machine.Spec.ProviderID
	if role == "master" {
		var err error
		deployment, err = newMasterDeployment(machine, cluster, release, 0, r.dnsServers, r.ntpServers)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	} else {
		var err error
		deployment, err = newWorkerDeployment(machine, cluster, release, r.dnsServers, r.ntpServers)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return deployment, nil
}

func (r *Resource) needsUpdate(ctx context.Context, cluster v1alpha2.KVMCluster, currentDeployment *v1.Deployment, desiredDeployment *v1.Deployment) (bool, error) {
	r.logger.Debugf(ctx, "creating Kubernetes client for tenant cluster")

	tcK8sClient, err := key.CreateK8sClientForTenantCluster(ctx, cluster, r.logger, r.tenantCluster)
	if tenantcluster.IsTimeout(err) {
		r.logger.Debugf(ctx, "did not create Kubernetes client for tenant cluster")
		r.logger.Debugf(ctx, "waiting for certificates timed out")

		return false, nil
	} else if tenant.IsAPINotAvailable(err) {
		r.logger.Debugf(ctx, "did not create Kubernetes client for tenant cluster")
		r.logger.Debugf(ctx, "tenant cluster is not available")

		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	// Updates can be quite disruptive. We have to be very careful with updating
	// resources that potentially imply disrupting customer workloads. We have
	// to check the state of all deployments before we can safely go ahead with
	// the update procedure.
	allReplicasUp := allNumbersEqual(currentDeployment.Status.AvailableReplicas, currentDeployment.Status.ReadyReplicas, currentDeployment.Status.Replicas, currentDeployment.Status.UpdatedReplicas)
	if !allReplicasUp {
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("cannot update any deployment: deployment '%s' must have all replicas up", currentDeployment.GetName()))
		return false, nil
	}

	if desiredDeployment == nil {
		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("not updating deployment '%s': no desired deployment found", currentDeployment.GetName()))
		return false, nil
	}

	if !isDeploymentModified(desiredDeployment, currentDeployment) {
		r.logger.Debugf(ctx, "not updating deployment '%s': no changes found", currentDeployment.GetName())
		return false, nil
	}

	// If worker deployment, check that master does not have any prohibited states before updating the worker
	if desiredDeployment.ObjectMeta.Labels[key.LabelApp] == key.WorkerID {
		tcNodes, err := tcK8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: "role=master"})
		if err != nil {
			return false, microerror.Mask(err)
		}
		for _, n := range tcNodes.Items {
			if key.NodeIsUnschedulable(n) {
				return false, nil
			}
		}
	}

	return true, nil
}
