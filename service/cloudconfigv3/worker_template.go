package cloudconfigv3

import (
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_2_0_0"
	"github.com/giantswarm/microerror"
)

type workerExtension struct {
	certs certs.Cluster
}

func (e *workerExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	filesMeta := []k8scloudconfig.FileMetadata{
		// Calico client.
		{
			AssetContent: string(e.certs.CalicoClient.CA),
			Path:         "/etc/kubernetes/ssl/calico/client-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs.CalicoClient.Crt),
			Path:         "/etc/kubernetes/ssl/calico/client-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs.CalicoClient.Key),
			Path:         "/etc/kubernetes/ssl/calico/client-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		// Etcd client.
		{
			AssetContent: string(e.certs.EtcdServer.CA),
			Path:         "/etc/kubernetes/ssl/etcd/client-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs.EtcdServer.Crt),
			Path:         "/etc/kubernetes/ssl/etcd/client-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs.EtcdServer.Key),
			Path:         "/etc/kubernetes/ssl/etcd/client-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		// Kubernetes worker.
		{
			AssetContent: string(e.certs.Worker.CA),
			Path:         "/etc/kubernetes/ssl/worker-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs.Worker.Crt),
			Path:         "/etc/kubernetes/ssl/worker-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs.Worker.Key),
			Path:         "/etc/kubernetes/ssl/worker-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
	}

	var newFiles []k8scloudconfig.FileAsset

	for _, fm := range filesMeta {
		c, err := k8scloudconfig.RenderAssetContent(fm.AssetContent, nil)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		fileAsset := k8scloudconfig.FileAsset{
			Metadata: fm,
			Content:  c,
		}

		newFiles = append(newFiles, fileAsset)
	}

	return newFiles, nil
}

func (e *workerExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		{
			Name:    "iscsid.service",
			Enable:  true,
			Command: "start",
		},
		{
			Name:    "multipathd.service",
			Enable:  true,
			Command: "start",
		},
	}

	var newUnits []k8scloudconfig.UnitAsset

	for _, fm := range unitsMeta {
		c, err := k8scloudconfig.RenderAssetContent(fm.AssetContent, nil)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		unitAsset := k8scloudconfig.UnitAsset{
			Metadata: fm,
			Content:  c,
		}

		newUnits = append(newUnits, unitAsset)
	}

	return newUnits, nil
}

func (e *workerExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	return nil
}
