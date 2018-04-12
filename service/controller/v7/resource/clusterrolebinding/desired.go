package clusterrolebinding

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/rbac/v1beta1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/v7/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "computing the new cluster role bindings")

	clusterRoleBindings, err := r.newClusterRoleBindings(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", fmt.Sprintf("computed the %d new cluster role bindings", len(clusterRoleBindings)))

	return clusterRoleBindings, nil
}

func (r *Resource) newClusterRoleBindings(customObject v1alpha1.KVMConfig) ([]*apiv1.ClusterRoleBinding, error) {
	var clusterRoleBindings []*apiv1.ClusterRoleBinding

	generalClusterRoleBinding := &apiv1.ClusterRoleBinding{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: apiv1.GroupName,
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: key.ClusterRoleBindingName(customObject),
			Labels: map[string]string{
				"app": "kvm-operator",
				"giantswarm.io/cluster-id":  key.ClusterID(customObject),
				"giantswarm.io/customer-id": key.ClusterCustomer(customObject),
			},
		},
		Subjects: []apiv1.Subject{
			{
				Kind:      apiv1.ServiceAccountKind,
				Namespace: key.ClusterID(customObject),
				Name:      key.ClusterID(customObject),
			},
		},
		RoleRef: apiv1.RoleRef{
			APIGroup: apiv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "kvm-operator",
		},
	}

	clusterRoleBindings = append(clusterRoleBindings, generalClusterRoleBinding)

	pspClusterRoleBinding := &apiv1.ClusterRoleBinding{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: apiv1.GroupName,
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name: key.ClusterRoleBindingPSPName(customObject),
			Labels: map[string]string{
				"app": "kvm-operator",
				"giantswarm.io/cluster-id":  key.ClusterID(customObject),
				"giantswarm.io/customer-id": key.ClusterCustomer(customObject),
			},
		},
		Subjects: []apiv1.Subject{
			{
				Kind:      apiv1.ServiceAccountKind,
				Namespace: key.ClusterID(customObject),
				Name:      key.ClusterID(customObject),
			},
		},
		RoleRef: apiv1.RoleRef{
			APIGroup: apiv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "kvm-operator-psp",
		},
	}

	clusterRoleBindings = append(clusterRoleBindings, pspClusterRoleBinding)

	return clusterRoleBindings, nil
}
