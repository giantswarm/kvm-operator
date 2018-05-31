package cloudconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_3_3_2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
)

// NewMasterTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewMasterTemplate(customObject v1alpha1.KVMConfig, certs certs.Cluster, node v1alpha1.ClusterNode, randomKeys randomkeys.Cluster) (string, error) {
	var err error

	var params k8scloudconfig.Params
	{
		params.APIServerEncryptionKey = string(randomKeys.APIServerEncryptionKey)
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &masterExtension{
			certs: certs,
		}
		params.Node = node
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		cloudConfigConfig := k8scloudconfig.DefaultCloudConfigConfig()
		cloudConfigConfig.Params = params
		cloudConfigConfig.Template = k8scloudconfig.MasterTemplate

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
	certs certs.Cluster
}

func (e *masterExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	var filesMeta []k8scloudconfig.FileMetadata

	for _, f := range certs.NewFilesClusterMaster(e.certs) {
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
			Enable:  true,
			Command: "start",
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
			Name:   "var-lib-etcd.mount",
			Enable: false,
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
		{
			AssetContent: `[Unit]
Description=Temporary fix for issues with calico-node and kube-proxy after master restart
Require=k8s-kubelet.service
After=k8s-kubelet.service

[Service]
Environment="KUBECONFIG=/etc/kubernetes/config/addons-kubeconfig.yml"
Environment="KUBECTL=quay.io/giantswarm/docker-kubectl:1dc536ec6dc4597ba46769b3d5d6ce53a7e62038"
ExecStart=/bin/sh -c "\
	while [ \"$(/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL get cs | grep Healthy | wc -l)\" -ne \"3\" ]; do sleep 1 && echo 'Waiting for healthy k8s'; done;sleep 30s; \
	/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system delete pod -l k8s-app=calico-node; \
	sleep 1m; \
	/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system delete pod -l k8s-app=kube-proxy; \
	sleep 1m; \
	/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system delete pod -l k8s-app=calico-kube-controllers; \
	/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system delete pod -l k8s-app=kube-dns"

[Install]
WantedBy=multi-user.target`,
			Name:    "calico-kube-kill.service",
			Command: "start",
			Enable:  true,
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
