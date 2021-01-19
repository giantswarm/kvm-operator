package clusterrolebinding

import (
	"context"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToKVMCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "computing the new cluster role bindings")

	clusterRoleBindings, err := r.newClusterRoleBindings(cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "computed the %d new cluster role bindings", len(clusterRoleBindings))

	return clusterRoleBindings, nil
}

func (r *Resource) newClusterRoleBindings(cr v1alpha2.KVMCluster) ([]*apiv1.ClusterRoleBinding, error) {
	var clusterRoleBindings []*apiv1.ClusterRoleBinding

	generalClusterRoleBinding := &apiv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: apiv1.GroupName,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: key.ClusterRoleBindingName(cr),
			Labels: map[string]string{
				"app":                       "kvm-operator",
				"giantswarm.io/cluster-id":  key.ClusterID(cr),
				"giantswarm.io/customer-id": key.ClusterCustomer(cr),
			},
		},
		Subjects: []apiv1.Subject{
			{
				Kind:      apiv1.ServiceAccountKind,
				Namespace: key.ClusterID(cr),
				Name:      key.ClusterID(cr),
			},
		},
		RoleRef: apiv1.RoleRef{
			APIGroup: apiv1.GroupName,
			Kind:     "ClusterRole",
			Name:     r.clusterRoleGeneral,
		},
	}

	clusterRoleBindings = append(clusterRoleBindings, generalClusterRoleBinding)

	pspClusterRoleBinding := &apiv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: apiv1.GroupName,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: key.ClusterRoleBindingPSPName(cr),
			Labels: map[string]string{
				"app":                       "kvm-operator",
				"giantswarm.io/cluster-id":  key.ClusterID(cr),
				"giantswarm.io/customer-id": key.ClusterCustomer(cr),
			},
		},
		Subjects: []apiv1.Subject{
			{
				Kind:      apiv1.ServiceAccountKind,
				Namespace: key.ClusterID(cr),
				Name:      key.ClusterID(cr),
			},
		},
		RoleRef: apiv1.RoleRef{
			APIGroup: apiv1.GroupName,
			Kind:     "ClusterRole",
			Name:     r.clusterRolePSP,
		},
	}

	clusterRoleBindings = append(clusterRoleBindings, pspClusterRoleBinding)

	return clusterRoleBindings, nil
}
