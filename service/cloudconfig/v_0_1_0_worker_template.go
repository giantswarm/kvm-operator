package cloudconfig

import (
	"github.com/giantswarm/certificatetpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/k8scloudconfig/v_0_1_0"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
)

func v_0_1_0WorkerTemplate(customObject kvmtpr.CustomObject, certs certificatetpr.AssetsBundle, node clustertprspec.Node) (string, error) {
	var err error

	var params v_0_1_0.Params
	{
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &v_0_1_0WorkerExtension{
			certs: certs,
		}
		params.Node = node
	}

	var newCloudConfig *v_0_1_0.CloudConfig
	{
		newCloudConfig, err = v_0_1_0.NewCloudConfig(v_0_1_0.WorkerTemplate, params)
		if err != nil {
			return "", microerror.Mask(err)
		}

		err = newCloudConfig.ExecuteTemplate()
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	return newCloudConfig.Base64(), nil
}

type v_0_1_0WorkerExtension struct {
	certs certificatetpr.AssetsBundle
}

func (e *v_0_1_0WorkerExtension) Files() ([]v_0_1_0.FileAsset, error) {
	filesMeta := []v_0_1_0.FileMetadata{
		// Calico client.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.CalicoComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/calico/client-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.CalicoComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/calico/client-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.CalicoComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/calico/client-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		// Etcd client.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/etcd/client-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/etcd/client-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/etcd/client-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		// Kubernetes worker.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.WorkerComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/worker-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.WorkerComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/worker-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.WorkerComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/worker-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
	}

	var newFiles []v_0_1_0.FileAsset

	for _, fm := range filesMeta {
		c, err := v_0_1_0.RenderAssetContent(fm.AssetContent, nil)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		fileAsset := v_0_1_0.FileAsset{
			Metadata: fm,
			Content:  c,
		}

		newFiles = append(newFiles, fileAsset)
	}

	return newFiles, nil
}

func (e *v_0_1_0WorkerExtension) Units() ([]v_0_1_0.UnitAsset, error) {
	return nil, nil
}

func (e *v_0_1_0WorkerExtension) VerbatimSections() []v_0_1_0.VerbatimSection {
	return nil
}
