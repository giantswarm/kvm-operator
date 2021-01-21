package cloudconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs/v3/pkg/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v10/pkg/template"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

type MasterConfig struct {
	Config Config
}

type Master struct {
	config Config
}

func NewMaster(config MasterConfig) (*Master, error) {
	err := config.Config.Validate()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	t := &Master{
		config: config.Config,
	}

	return t, nil
}

// NewMasterTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *Master) NewTemplate(ctx context.Context, cr v1alpha2.KVMCluster, data IgnitionTemplateData, nodeIndex int) (string, error) {
	var extension *masterExtension
	{
		certFiles, err := fetchCertFiles(ctx, data.CertsSearcher, key.ClusterID(&cr), masterCertFiles)
		if err != nil {
			return "", microerror.Mask(err)
		}
		extension = &masterExtension{
			certs:     certFiles,
			cr:        cr,
			nodeIndex: nodeIndex,
		}
	}

	var apiExtraArgs []string
	{
		if key.OIDCClientID(cr) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-client-id=%s", key.OIDCClientID(cr)))
		}
		if key.OIDCIssuerURL(cr) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-issuer-url=%s", key.OIDCIssuerURL(cr)))
		}
		if key.OIDCUsernameClaim(cr) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-username-claim=%s", key.OIDCUsernameClaim(cr)))
		}
		if key.OIDCGroupsClaim(cr) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-groups-claim=%s", key.OIDCGroupsClaim(cr)))
		}

		apiExtraArgs = append(apiExtraArgs, c.config.APIExtraArgs...)
	}

	var kubeletExtraArgs []string
	{
		if c.config.PodInfraContainerImage != "" {
			kubeletExtraArgs = append(kubeletExtraArgs, fmt.Sprintf("--pod-infra-container-image=%s", c.config.PodInfraContainerImage))
		}

		kubeletExtraArgs = append(kubeletExtraArgs, c.config.KubeletExtraArgs...)
	}

	var params k8scloudconfig.Params
	{
		params.APIServerEncryptionKey = string(data.ClusterKeys.APIServerEncryptionKey)
		params.BaseDomain = key.BaseDomain(cr)
		params.Cluster = clusterToLegacy(c.config, cr, "").Cluster
		// Ingress controller service remains in k8scloudconfig and will be
		// removed in a later migration.
		params.DisableIngressControllerService = true
		params.Etcd = k8scloudconfig.Etcd{
			ClientPort:          key.EtcdPort,
			HighAvailability:    false,
			InitialClusterState: k8scloudconfig.InitialClusterStateNew,
		}
		params.Extension = extension
		params.ImagePullProgressDeadline = key.DefaultImagePullProgressDeadline
		params.Kubernetes.Apiserver.CommandExtraArgs = apiExtraArgs
		params.Images = data.Images
		params.Versions = data.Versions
		params.RegistryMirrors = c.config.RegistryMirrors
		params.SSOPublicKey = c.config.SSOPublicKey
		params.DockerhubToken = c.config.DockerhubToken
		params.Kubernetes.Kubelet.CommandExtraArgs = kubeletExtraArgs

		ignitionPath := k8scloudconfig.GetIgnitionPath(c.config.IgnitionPath)
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
			Template: k8scloudconfig.MasterTemplate,
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

type masterExtension struct {
	certs     []certs.File
	cr        v1alpha2.KVMCluster
	nodeIndex int
}

func (e *masterExtension) Files() ([]k8scloudconfig.FileAsset, error) {
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
		AssetContent: fmt.Sprintf("InitiatorName=%s", key.IscsiInitiatorName(e.cr, e.nodeIndex, key.MasterID)),
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

	calicoKubeKillFile := k8scloudconfig.FileMetadata{
		AssetContent: calicoKubeKillScript,
		Path:         "/opt/calico-kube-kill",
		Owner: k8scloudconfig.Owner{
			User: k8scloudconfig.User{
				Name: FileOwnerUserName,
			},
			Group: k8scloudconfig.Group{
				Name: FileOwnerGroupName,
			},
		},
		Permissions: 0755,
	}
	filesMeta = append(filesMeta, calicoKubeKillFile)

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

func (e *masterExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		// Mount etcd volume when directory first accessed
		// This automount is workaround for
		// https://bugzilla.redhat.com/show_bug.cgi?id=1184122
		{
			AssetContent: `[Unit]
Description=Automount for etcd volume
[Automount]
Where=/var/lib/etcd
[Install]
WantedBy=multi-user.target
`,
			Name:    "var-lib-etcd.automount",
			Enabled: true,
		},
		// Mount for etcd volume activated by automount
		{
			AssetContent: `[Unit]
Description=Mount for etcd volume
[Mount]
What=etcdshare
Where=/var/lib/etcd
Options=trans=virtio,version=9p2000.L,cache=mmap
Type=9p
[Install]
WantedBy=multi-user.target
`,
			Name:    "var-lib-etcd.mount",
			Enabled: false,
		},
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
		{
			AssetContent: `[Unit]
Description=Temporary fix for issues with calico-node and kube-proxy after master restart
Requires=k8s-kubelet.service
After=k8s-kubelet.service

[Service]
Environment="PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/opt/bin"
Environment="KUBECONFIG=/etc/kubernetes/kubeconfig/addons.yaml"
ExecStart=/opt/calico-kube-kill

[Install]
WantedBy=multi-user.target`,
			Name:    "calico-kube-kill.service",
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

func (e *masterExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	return nil
}
