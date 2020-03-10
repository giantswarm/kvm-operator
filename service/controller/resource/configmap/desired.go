package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/controllercontext"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing the new config maps")

	configMaps, err := r.newConfigMaps(cr, cc.Spec.Versions)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed the %d new config maps", len(configMaps)))

	return configMaps, nil
}

func (r *Resource) newConfigMaps(cr v1alpha1.KVMConfig, versions controllercontext.ComponentVersions) ([]*corev1.ConfigMap, error) {
	var configMaps []*corev1.ConfigMap

	certs, err := r.certsSearcher.SearchCluster(key.ClusterID(cr))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	keys, err := r.keyWatcher.SearchCluster(key.ClusterID(cr))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, node := range cr.Spec.Cluster.Masters {
		nodeIdx, exists := key.NodeIndex(cr, node.ID)
		if !exists {
			return nil, microerror.Maskf(notFoundError, fmt.Sprintf("node index for master (%q) is not available", node.ID))
		}

		params := cloudconfig.MasterTemplateParams{
			CR:         cr,
			Certs:      certs,
			Node:       node,
			RandomKeys: keys,
			NodeIndex:  nodeIdx,
			Versions:   nil,
		}
		template, err := r.cloudConfig.NewMasterTemplate(params)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMap, err := r.newConfigMap(cr, template, node, key.MasterID)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	for _, node := range customResource.Spec.Cluster.Workers {
		nodeIdx, exists := key.NodeIndex(customResource, node.ID)
		if !exists {
			return nil, microerror.Maskf(notFoundError, fmt.Sprintf("node index for worker (%q) is not available", node.ID))
		}

		template, err := r.cloudConfig.NewWorkerTemplate(customResource, certs, node, nodeIdx)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMap, err := r.newConfigMap(customResource, template, node, key.WorkerID)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

// newConfigMap creates a new Kubernetes configmap using the provided
// information. customResource is used for name and label creation. params
// serves as structure being injected into the template execution to interpolate
// variables. prefix can be either "master" or "worker" and is used to prefix
// the configmap name.
func (r *Resource) newConfigMap(customResource v1alpha1.KVMConfig, template string, node v1alpha1.ClusterNode, prefix string) (*corev1.ConfigMap, error) {
	var newConfigMap *corev1.ConfigMap
	{
		newConfigMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: key.ConfigMapName(customResource, node, prefix),
				Labels: map[string]string{
					// TODO: Delete two legacy labels from next release
					// issues: https://github.com/giantswarm/giantswarm/issues/7771
					"cluster":          key.ClusterID(customResource),
					"customer":         key.ClusterCustomer(customResource),
					label.Cluster:      key.ClusterID(customResource),
					label.ManagedBy:    project.Name(),
					label.Organization: key.ClusterCustomer(customResource),
				},
			},
			Data: map[string]string{
				KeyUserData: template,
			},
		}
	}

	return newConfigMap, nil
}
