package cloudconfig

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/k8scloudconfig/v6/pkg/template"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

// NewWorkerTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewWorkerTemplate(customObject v1alpha1.KVMConfig, data IgnitionTemplateData, node v1alpha1.ClusterNode, nodeIndex int) (string, error) {
	var err error

	var params template.Params
	{
		params = template.DefaultParams()

		params.BaseDomain = key.BaseDomain(customObject)
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &workerExtension{
			certs:        data.ClusterCerts,
			customObject: customObject,
			nodeIndex:    nodeIndex,
		}
		params.Images = data.Images
		params.Node = node
		params.SSOPublicKey = c.ssoPublicKey

		ignitionPath := template.GetIgnitionPath(c.ignitionPath)
		params.Files, err = template.RenderFiles(ignitionPath, params)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	var newCloudConfig *template.CloudConfig
	{
		cloudConfigConfig := template.DefaultCloudConfigConfig()
		cloudConfigConfig.Params = params
		cloudConfigConfig.Template = template.WorkerTemplate

		newCloudConfig, err = template.NewCloudConfig(cloudConfigConfig)
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
	certs        certs.Cluster
	customObject v1alpha1.KVMConfig
	nodeIndex    int
}

func (e *workerExtension) Files() ([]template.FileAsset, error) {
	var filesMeta []template.FileMetadata

	for _, f := range certs.NewFilesClusterWorker(e.certs) {
		m := template.FileMetadata{
			AssetContent: string(f.Data),
			Path:         f.AbsolutePath,
			Owner: template.Owner{
				User: template.User{
					Name: FileOwnerUserName,
				},
				Group: template.Group{
					Name: FileOwnerGroupName,
				},
			},
			Permissions: FilePermission,
		}
		filesMeta = append(filesMeta, m)
	}

	iscsiInitiatorFile := template.FileMetadata{
		AssetContent: fmt.Sprintf("InitiatorName=%s", key.IscsiInitiatorName(e.customObject, e.nodeIndex, key.WorkerID)),
		Path:         IscsiInitiatorNameFilePath,
		Owner: template.Owner{
			User: template.User{
				Name: FileOwnerUserName,
			},
			Group: template.Group{
				Name: FileOwnerGroupName,
			},
		},
		Permissions: IscsiInitiatorFilePermissions,
	}
	filesMeta = append(filesMeta, iscsiInitiatorFile)

	// iscsi config as workaround for this bug https://github.com/kubernetes/kubernetes/issues/73181
	iscsiConfigFile := template.FileMetadata{
		AssetContent: IscsiConfigFileContent,
		Path:         IscsiConfigFilePath,
		Owner: template.Owner{
			User: template.User{
				Name: FileOwnerUserName,
			},
			Group: template.Group{
				Name: FileOwnerGroupName,
			},
		},
		Permissions: IscsiConfigFilePermissions,
	}
	filesMeta = append(filesMeta, iscsiConfigFile)

	var newFiles []template.FileAsset

	for _, fm := range filesMeta {
		c, err := template.RenderFileAssetContent(fm.AssetContent, nil)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		fileAsset := template.FileAsset{
			Metadata: fm,
			Content:  c,
		}

		newFiles = append(newFiles, fileAsset)
	}

	return newFiles, nil
}

func (e *workerExtension) Units() ([]template.UnitAsset, error) {
	unitsMeta := []template.UnitMetadata{
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
			Enabled: true,
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
			Enabled: true,
		},
		{
			Name:    "iscsid.service",
			Enabled: true,
		},
		{
			Name:    "multipathd.service",
			Enabled: true,
		},
	}

	var newUnits []template.UnitAsset

	for _, fm := range unitsMeta {
		c, err := template.RenderAssetContent(fm.AssetContent, nil)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		unitAsset := template.UnitAsset{
			Metadata: fm,
			Content:  c,
		}

		newUnits = append(newUnits, unitAsset)
	}

	return newUnits, nil
}

func (e *workerExtension) VerbatimSections() []template.VerbatimSection {
	return nil
}
