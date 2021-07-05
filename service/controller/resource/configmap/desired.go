package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v10/pkg/template"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/v4/pkg/label"
	"github.com/giantswarm/kvm-operator/v4/pkg/project"
	"github.com/giantswarm/kvm-operator/v4/service/controller/cloudconfig"
	"github.com/giantswarm/kvm-operator/v4/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customResource, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "computing the new config maps")

	configMaps, err := r.newConfigMaps(ctx, customResource)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "computed the %d new config maps", len(configMaps))

	return configMaps, nil
}

func (r *Resource) newConfigMaps(ctx context.Context, customResource v1alpha1.KVMConfig) ([]*corev1.ConfigMap, error) {
	var configMaps []*corev1.ConfigMap

	keys, err := r.keyWatcher.SearchCluster(ctx, key.ClusterID(customResource))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var release releasev1alpha1.Release
	{
		releaseVersion := customResource.Labels[label.ReleaseVersion]
		releaseName := fmt.Sprintf("v%s", releaseVersion)
		err = r.ctrlClient.Get(ctx, client.ObjectKey{Name: releaseName}, &release)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	versions, err := k8scloudconfig.ExtractComponentVersions(release.Spec.Components)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	defaultVersions := key.DefaultVersions()
	versions.KubernetesAPIHealthz = defaultVersions.KubernetesAPIHealthz
	versions.KubernetesNetworkSetupDocker = defaultVersions.KubernetesNetworkSetupDocker
	images := k8scloudconfig.BuildImages(r.registryDomain, versions)

	data := cloudconfig.IgnitionTemplateData{
		CustomObject:  customResource,
		CertsSearcher: r.certsSearcher,
		ClusterKeys:   keys,
		Images:        images,
		Versions:      versions,
	}

	for _, node := range customResource.Spec.Cluster.Masters {
		nodeIdx, exists := key.NodeIndex(customResource, node.ID)
		if !exists {
			return nil, microerror.Maskf(notFoundError, fmt.Sprintf("node index for master (%q) is not available", node.ID))
		}

		template, err := r.cloudConfig.NewMasterTemplate(ctx, customResource, data, node, nodeIdx)
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

		template, err := r.cloudConfig.NewWorkerTemplate(ctx, customResource, data, node, nodeIdx)
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
