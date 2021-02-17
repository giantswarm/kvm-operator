package deployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/crud"
	"github.com/giantswarm/tenantcluster/v4/pkg/tenantcluster"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	deploymentsToUpdate, err := toDeployments(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(deploymentsToUpdate) != 0 {
		r.logger.Debugf(ctx, "updating the deployments in the Kubernetes API")

		namespace := key.ClusterNamespace(customResource)
		for _, deployment := range deploymentsToUpdate {
			_, err := r.k8sClient.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Debugf(ctx, "updated the deployments in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the deployments do not need to be updated in the Kubernetes API")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	delete, err := r.newDeleteChangeForUpdatePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetCreateChange(create)
	patch.SetDeleteChange(delete)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "creating Kubernetes client for tenant cluster")

	tcK8sClient, err := key.CreateK8sClientForWorkloadCluster(ctx, cr, r.logger, r.tenantCluster)
	if tenantcluster.IsTimeout(err) {
		r.logger.Debugf(ctx, "did not create Kubernetes client for tenant cluster")
		r.logger.Debugf(ctx, "waiting for certificates timed out")

		return nil, nil
	} else if tenant.IsAPINotAvailable(err) {
		r.logger.Debugf(ctx, "did not create Kubernetes client for tenant cluster")
		r.logger.Debugf(ctx, "tenant cluster is not available")

		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "created Kubernetes client for tenant cluster")

	return r.updateDeployments(ctx, currentState, desiredState, tcK8sClient.CtrlClient())
}

func (r *Resource) updateDeployments(ctx context.Context, currentState, desiredState interface{}, tcK8sClient client.Client) (interface{}, error) {
	currentDeployments, err := toDeployments(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredDeployments, err := toDeployments(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out which deployments have to be updated")

	// Updates can be quite disruptive. We have to be very careful with updating
	// resources that potentially imply disrupting customer workloads. We have
	// to check the state of all deployments before we can safely go ahead with
	// the update procedure.
	for _, d := range currentDeployments {
		allReplicasUp := allNumbersEqual(d.Status.AvailableReplicas, d.Status.ReadyReplicas, d.Status.Replicas, d.Status.UpdatedReplicas)
		if !allReplicasUp {
			r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("cannot update any deployment: deployment '%s' must have all replicas up", d.GetName()))
			return nil, nil
		}
	}

	// We select one deployment to be updated per reconciliation loop. Therefore
	// we have to check its state on the version bundle level to see if a
	// deployment is already up to date. We also check if there are any other
	// changes on the pod specs. In case there are none, we check the next one.
	// The first one not being up to date will be chosen to be updated next and
	// the loop will be broken immediately.
DeploymentsLoop:
	for _, currentDeployment := range currentDeployments {
		desiredDeployment, err := getDeploymentByName(desiredDeployments, currentDeployment.Name)
		if IsNotFound(err) {
			// NOTE that this case indicates we should remove the current deployment
			// eventually.
			r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("not updating deployment '%s': no desired deployment found", currentDeployment.GetName()))
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if !isDeploymentModified(desiredDeployment, currentDeployment) {
			r.logger.Debugf(ctx, "not updating deployment '%s': no changes found", currentDeployment.GetName())
			continue
		}

		// If worker deployment, check that master does not have any prohibited states before updating the worker
		if desiredDeployment.ObjectMeta.Labels[key.LabelApp] == key.WorkerID {
			// List all master nodes in the tenant
			tcNodes, err := tcK8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: "role=master"})
			if err != nil {
				r.logger.LogCtx(ctx, "level", "warning", "message", "unable to list tenant cluster master nodes")
				return nil, microerror.Mask(err)
			}
			for _, n := range tcNodes.Items {
				if key.NodeIsUnschedulable(n) {
					// Node has NoSchedule or NoExecute taint
					msg := fmt.Sprintf("not updating deployment '%s': one or more tenant cluster master nodes are unschedulable", currentDeployment.GetName())
					r.logger.LogCtx(ctx, "level", "warning", "message", msg)
					continue DeploymentsLoop
				}
			}
		}

		r.logger.Debugf(ctx, "found deployment '%s' that has to be updated", desiredDeployment.GetName())

		return []*v1.Deployment{desiredDeployment}, nil
	}

	return nil, nil
}
