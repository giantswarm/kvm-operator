package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller/cloudconfig"
	"github.com/giantswarm/kvm-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "computing the new config maps")

	configMaps, err := r.newConfigMaps(customResource)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("computed the %d new config maps", len(configMaps)))

	return configMaps, nil
}

func (r *Resource) newConfigMaps(customResource v1alpha1.KVMConfig) ([]*corev1.ConfigMap, error) {
	var configMaps []*corev1.ConfigMap

	certs, err := r.certsSearcher.SearchCluster(key.ClusterID(customResource))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	keys, err := r.keyWatcher.SearchCluster(key.ClusterID(customResource))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var release *releasev1alpha1.Release
	{
		releaseVersion := customResource.Labels[label.ReleaseVersion]
		releaseName := fmt.Sprintf("v%s", releaseVersion)
		release, err = r.g8sClient.ReleaseV1alpha1().Releases().Get(releaseName, metav1.GetOptions{})
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	versions, err := v_6_0_0.ExtractComponentVersions(release.Spec.Components)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	defaultVersions := key.DefaultVersions()
	versions.Kubectl = defaultVersions.Kubectl
	versions.KubernetesAPIHealthz = defaultVersions.KubernetesAPIHealthz
	versions.KubernetesNetworkSetupDocker = defaultVersions.KubernetesNetworkSetupDocker
	images := v_6_0_0.BuildImages(r.registryDomain, versions)

	data := cloudconfig.IgnitionTemplateData{
		CustomObject: customResource,
		ClusterCerts: certs,
		ClusterKeys:  keys,
		Images:       images,
	}

	for _, node := range customResource.Spec.Cluster.Masters {
		nodeIdx, exists := key.NodeIndex(customResource, node.ID)
		if !exists {
			return nil, microerror.Maskf(notFoundError, fmt.Sprintf("node index for master (%q) is not available", node.ID))
		}

		template, err := r.cloudConfig.NewMasterTemplate(customResource, data, node, nodeIdx)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMap, err := r.newConfigMap(customResource, template, node, key.MasterID)
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

		template, err := r.cloudConfig.NewWorkerTemplate(customResource, data, node, nodeIdx)
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
