package cloudconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/v3/pkg/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v10/pkg/template"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

// NewWorkerTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewWorkerTemplate(ctx context.Context, cr v1alpha2.KVMCluster, data IgnitionTemplateData, node v1alpha1.ClusterNode, nodeIndex int) (string, error) {
	var extension *workerExtension
	{
		certFiles, err := fetchCertFiles(ctx, data.CertsSearcher, key.ClusterID(cr), workerCertFiles)
		if err != nil {
			return "", microerror.Mask(err)
		}
		extension = &workerExtension{
			certs:     certFiles,
			cr:        cr,
			nodeIndex: nodeIndex,
		}
	}

	var params k8scloudconfig.Params
	{
		params.BaseDomain = key.BaseDomain(cr)
		params.Cluster = cr.Spec.Cluster
		params.Extension = extension
		params.Images = data.Images
		params.Versions = data.Versions
		params.Node = node
		params.RegistryMirrors = c.registryMirrors
		params.SSOPublicKey = c.ssoPublicKey
		params.ImagePullProgressDeadline = key.DefaultImagePullProgressDeadline
		params.DockerhubToken = c.dockerhubToken

		ignitionPath := k8scloudconfig.GetIgnitionPath(c.ignitionPath)
		{
			var err error
			params.Files, err = k8scloudconfig.RenderFiles(ignitionPath, params)
			if err != nil {
				return "", microerror.Mask(err)
			}
		}
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		cloudConfigConfig := k8scloudconfig.CloudConfigConfig{
			Params:   params,
			Template: k8scloudconfig.WorkerTemplate,
		}

		var err error
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
	certs     []certs.File
	cr        v1alpha2.KVMCluster
	nodeIndex int
}

func (e *workerExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	var filesMeta []k8scloudconfig.FileMetadata

	for _, f := range e.certs {
		m := k8scloudconfig.FileMetadata{
			AssetContent: string(f.Data),
			Path:         f.AbsolutePath,
			Owner: k8scloudconfig.Owner{
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
			},
			Permissions: FilePermission,
		}
		filesMeta = append(filesMeta, m)
	}

	iscsiInitiatorFile := k8scloudconfig.FileMetadata{
		AssetContent: fmt.Sprintf("InitiatorName=%s", key.IscsiInitiatorName(e.cr, e.nodeIndex, key.WorkerID)),
		Path:         IscsiInitiatorNameFilePath,
		Owner: k8scloudconfig.Owner{
			User: k8scloudconfig.User{
				Name: FileOwnerUserName,
			},
			Group: k8scloudconfig.Group{
				Name: FileOwnerGroupName,
			},
		},
		Permissions: IscsiInitiatorFilePermissions,
	}
	filesMeta = append(filesMeta, iscsiInitiatorFile)

	// iscsi config as workaround for this bug https://github.com/kubernetes/kubernetes/issues/73181
	iscsiConfigFile := k8scloudconfig.FileMetadata{
		AssetContent: IscsiConfigFileContent,
		Path:         IscsiConfigFilePath,
		Owner: k8scloudconfig.Owner{
			User: k8scloudconfig.User{
				Name: FileOwnerUserName,
			},
			Group: k8scloudconfig.Group{
				Name: FileOwnerGroupName,
			},
		},
		Permissions: IscsiConfigFilePermissions,
	}
	filesMeta = append(filesMeta, iscsiConfigFile)

	var newFiles []k8scloudconfig.FileAsset

	for _, fm := range filesMeta {
		c, err := k8scloudconfig.RenderFileAssetContent(fm.AssetContent, nil)
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
