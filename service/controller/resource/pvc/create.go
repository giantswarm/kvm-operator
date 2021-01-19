package pvc

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cr, err := key.ToKVMCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	pvcsToCreate, err := toPVCs(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(pvcsToCreate) != 0 {
		r.logger.Debugf(ctx, "creating the PVCs in the Kubernetes API")

		namespace := key.ClusterNamespace(cr)
		for _, PVC := range pvcsToCreate {
			_, err := r.k8sClient.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, PVC, v1.CreateOptions{})
			if apierrors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Debugf(ctx, "created the PVCs in the Kubernetes API")
	} else {
		r.logger.Debugf(ctx, "the PVCs do not need to be created in the Kubernetes API")
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentPVCs, err := toPVCs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredPVCs, err := toPVCs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out which PVCs have to be created")

	var pvcsToCreate []*corev1.PersistentVolumeClaim

	for _, desiredPVC := range desiredPVCs {
		if !containsPVC(currentPVCs, desiredPVC) {
			pvcsToCreate = append(pvcsToCreate, desiredPVC)
		}
	}

	r.logger.Debugf(ctx, "found %d PVCs that have to be created", len(pvcsToCreate))

	return pvcsToCreate, nil
}
