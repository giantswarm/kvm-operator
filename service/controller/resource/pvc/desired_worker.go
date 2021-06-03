package pvc

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/to"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) getDesiredWorkerPVCs(ctx context.Context, customObject v1alpha1.KVMConfig) ([]corev1.PersistentVolumeClaim, error) {
	var persistentVolumeClaims []corev1.PersistentVolumeClaim
	namespace := key.ClusterNamespace(customObject)

	for i, workerKVM := range customObject.Spec.KVM.Workers {
		workerCluster := customObject.Spec.Cluster.Workers[i]

		if key.IsDeleted(&customObject) && len(workerKVM.HostVolumes) > 0 {
			deploymentName := key.DeploymentName(key.WorkerID, workerCluster.ID)
			_, err := r.k8sClient.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				// the cluster is being deleted and the worker deployment doesn't exist
				r.logger.Debugf(ctx, "worker deployment %#q not found, not adding pvcs to desired state", deploymentName)
			} else if err != nil {
				return nil, microerror.Mask(err)
			} else {
				// deployment still exists, keep finalizer, cancel, and delete on next reconciliation
				r.logger.Debugf(ctx, "keeping finalizer and canceling as deployment %#q still exists", deploymentName)
				finalizerskeptcontext.SetKept(ctx)
				resourcecanceledcontext.SetCanceled(ctx)
				return nil, nil
			}
		}

		for _, hostVolume := range workerKVM.HostVolumes {
			candidateVolumes, err := r.k8sClient.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{
				LabelSelector: labels.FormatLabels(map[string]string{
					key.LabelMountTag: hostVolume.MountTag,
				}),
			})
			if err != nil {
				return nil, microerror.Mask(err)
			} else if len(candidateVolumes.Items) == 0 {
				return nil, microerror.Maskf(notFoundError, "persistent volume with mount tag %#q not found", hostVolume.MountTag)
			} else if len(candidateVolumes.Items) > 1 {
				return nil, microerror.Maskf(notFoundError, "multiple persistent volumes with mount tag %#q found", hostVolume.MountTag)
			}

			pv := candidateVolumes.Items[0]
			claimName := key.LocalWorkerPVCName(key.ClusterID(customObject), key.VMNumber(i), hostVolume.MountTag)

			if pv.Spec.ClaimRef != nil && (pv.Spec.ClaimRef.Namespace != namespace || pv.Spec.ClaimRef.Name != claimName) {
				// PV already bound to a different PVC
				return nil, microerror.Maskf(isAlreadyBound, "persistent volume %#q is already bound to %#q", pv.Name, pv.Spec.ClaimRef.Name)
			}

			persistentVolumeClaim := corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: claimName,
					Labels: map[string]string{
						key.LabelCustomer:      key.ClusterCustomer(customObject),
						key.LabelApp:           key.WorkerID,
						key.LabelCluster:       key.ClusterID(customObject),
						key.LegacyLabelCluster: key.ClusterID(customObject),
						key.LabelVersionBundle: key.OperatorVersion(customObject),
						"node":                 workerCluster.ID,
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: pv.Spec.Capacity,
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							key.LabelMountTag: hostVolume.MountTag,
						},
					},
					StorageClassName: to.StringP(LocalStorageClass),
				},
			}

			persistentVolumeClaims = append(persistentVolumeClaims, persistentVolumeClaim)
		}
	}

	return persistentVolumeClaims, nil
}
