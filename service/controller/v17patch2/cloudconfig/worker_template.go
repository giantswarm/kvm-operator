package cloudconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_3_7_3"
	"github.com/giantswarm/microerror"
)

// NewWorkerTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewWorkerTemplate(customObject v1alpha1.KVMConfig, certs certs.Cluster, node v1alpha1.ClusterNode) (string, error) {
	var err error

	var params k8scloudconfig.Params
	{
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &workerExtension{
			certs: certs,
		}
		params.Node = node
		params.SSOPublicKey = c.ssoPublicKey
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		cloudConfigConfig := k8scloudconfig.DefaultCloudConfigConfig()
		cloudConfigConfig.Params = params
		cloudConfigConfig.Template = k8scloudconfig.WorkerTemplate

		newCloudConfig, err = k8scloudconfig.NewCloudConfig(cloudConfigConfig)
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

type workerExtension struct {
	certs certs.Cluster
}

func (e *workerExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	var filesMeta []k8scloudconfig.FileMetadata

	for _, f := range certs.NewFilesClusterWorker(e.certs) {
		m := k8scloudconfig.FileMetadata{
			AssetContent: string(f.Data),
			Path:         f.AbsolutePath,
			Owner:        FileOwner,
			Permissions:  FilePermission,
		}
		filesMeta = append(filesMeta, m)
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
			AssetContent: `[Unit]
Before=docker.service
Description=Mount for docker volume
[Mount]
What=/dev/disk/by-id/virtio-dockerfs
Where=/var/lib/docker
Type=xfs
[Install]
WantedBy=multi-user.target
`,
			Name:    "var-lib-docker.mount",
			Enable:  true,
			Command: "start",
		},
		{
			AssetContent: `[Unit]
Before=docker.service
Description=Mount for kubelet volume
[Mount]
What=/dev/disk/by-id/virtio-kubeletfs
Where=/var/lib/kubelet
Type=xfs
[Install]
WantedBy=multi-user.target
`,
			Name:    "var-lib-kubelet.mount",
			Enable:  true,
			Command: "start",
		},
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
