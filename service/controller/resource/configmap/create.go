package configmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v10/pkg/template"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/kvm-operator/pkg/label"
	"github.com/giantswarm/kvm-operator/pkg/project"
	"github.com/giantswarm/kvm-operator/service/controller/cloudconfig"
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
		err := r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{
			Namespace: clusterID,
			Name:      clusterID,
		}, &kvmCluster)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var existing corev1.ConfigMap
	err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{
		Namespace: clusterID,
		Name:      clusterID,
	}, &existing)
	if apierrors.IsNotFound(err) {
		toCreate, err := r.newConfigMap(ctx, kvmCluster, cr)
		if err != nil {
			return microerror.Mask(err)
		}
		err = r.k8sClient.CtrlClient().Create(ctx, toCreate)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) newConfigMap(ctx context.Context, cluster v1alpha2.KVMCluster, machine v1alpha2.KVMMachine) (*corev1.ConfigMap, error) {
	clusterID := machine.Namespace
	keys, err := r.keyWatcher.SearchCluster(ctx, clusterID)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var release releasev1alpha1.Release
	{
		releaseVersion := machine.Labels[label.ReleaseVersion]
		var release releasev1alpha1.Release
		err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{
			Name: fmt.Sprintf("v%s", releaseVersion),
		}, &release)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var data cloudconfig.IgnitionTemplateData
	{
		versions, err := k8scloudconfig.ExtractComponentVersions(release.Spec.Components)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		defaultVersions := key.DefaultVersions()
		versions.KubernetesAPIHealthz = defaultVersions.KubernetesAPIHealthz
		versions.KubernetesNetworkSetupDocker = defaultVersions.KubernetesNetworkSetupDocker
		images := k8scloudconfig.BuildImages(r.registryDomain, versions)
		data = cloudconfig.IgnitionTemplateData{
			CustomResource: machine,
			CertsSearcher:  r.certsSearcher,
			ClusterKeys:    keys,
			Images:         images,
			Versions:       versions,
		}
	}

	role := key.WorkerID
	if machine.Labels["cluster.x-k8s.io/control-plane"] == "true" {
		role = key.MasterID
	}

	nodeIdx, exists := key.NodeIndex(cluster, machine.Spec.ProviderID)
	if !exists {
		return nil, microerror.Maskf(notFoundError, fmt.Sprintf("node index for %s (%q) is not available", role, machine.Spec.ProviderID))
	}

	var commonConfig cloudconfig.Config
	{
		commonConfig = cloudconfig.Config{
			CertsSearcher:      r.certsSearcher,
			K8sClient:          r.k8sClient,
			Logger:             r.logger,
			RandomKeysSearcher: r.keyWatcher,
			OIDC: cloudconfig.OIDCConfig{
				ClientID:      cluster.Spec.Cluster.OIDC.ClientID,
				IssuerURL:     cluster.Spec.Cluster.OIDC.IssuerURL,
				UsernameClaim: cluster.Spec.Cluster.OIDC.Claims.Username,
				GroupsClaim:   cluster.Spec.Cluster.OIDC.Claims.Groups,
			},
			DockerhubToken:  r.dockerhubToken,
			IgnitionPath:    r.ignitionPath,
			RegistryDomain:  r.registryDomain,
			RegistryMirrors: r.registryMirrors,
		}
	}

	var template string
	if role == "master" {
		config := cloudconfig.MasterConfig{Config: commonConfig}
		cloudConfig, err := cloudconfig.NewMaster(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		template, err = cloudConfig.NewTemplate(ctx, cluster, data, nodeIdx)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	} else {
		config := cloudconfig.WorkerConfig{Config: commonConfig}
		cloudConfig, err := cloudconfig.NewWorker(config)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		template, err = cloudConfig.NewTemplate(ctx, cluster, data, nodeIdx)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.ConfigMapName(machine, role),
			Namespace: key.ClusterNamespace(&cluster),
			Labels: map[string]string{
				label.Cluster:      key.ClusterID(&cluster),
				label.ManagedBy:    project.Name(),
				label.Organization: key.ClusterCustomer(&cluster),
			},
		},
		Data: map[string]string{
			KeyUserData: template,
		},
	}, nil
}
