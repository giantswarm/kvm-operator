package cloudconfigv2

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certificatetpr"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_2_0_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"
)

type v2_0_0masterExtension struct {
	certs certificatetpr.AssetsBundle
	keys  map[randomkeytpr.Key][]byte
}

// NewMasterTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func v2_0_0MasterTemplate(customObject v1alpha1.KVMConfig, certs certificatetpr.AssetsBundle, node v1alpha1.ClusterNode, keys map[randomkeytpr.Key][]byte) (string, error) {
	var err error

	_, ok := keys[randomkeytpr.EncryptionKey]
	if !ok {
		return "", microerror.Maskf(notFoundError, "could not get encryption keys from secrets")
	}

	var params k8scloudconfig.Params
	{
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &v2_0_0masterExtension{
			certs: certs,
			keys:  keys,
		}
		params.Node = node
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		newCloudConfig, err = k8scloudconfig.NewCloudConfig(k8scloudconfig.MasterTemplate, params)
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

func (e *v2_0_0masterExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	encryptionConfig, err := EncryptionConfig(string(e.keys[randomkeytpr.EncryptionKey]))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	filesMeta := []k8scloudconfig.FileMetadata{
		// Kubernetes API server.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.APIComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/apiserver-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.APIComponent, certificatetpr.Crt}]),
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.APIComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/apiserver-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
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
		// Etcd server.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/etcd/server-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/etcd/server-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/etcd/server-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		// Service account.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.ServiceAccountComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/service-account-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.ServiceAccountComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/service-account-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.ServiceAccountComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/service-account-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		// Encryption key
		{
			AssetContent: encryptionConfig,
			Path:         "/etc/kubernetes/encryption/k8s-encryption-config.yaml",
			Owner:        FileOwner,
			Permissions:  0600,
		},
		// etcd_data_dir drop-in
		{
			AssetContent: etcd_data_dir_dropin,
			Path:         "/etc/systemd/system/etc-kubernetes-data-etcd.mount.d/00-before-set-ownership.conf",
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

func (e *v2_0_0masterExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		// Mount etcd volume when directory first accessed
		// This automount is workaround for
		// https://bugzilla.redhat.com/show_bug.cgi?id=1184122
		{
			AssetContent: `[Unit]
Description=Automount for etcd volume
[Automount]
Where=/etc/kubernetes/data/etcd
[Install]
WantedBy=multi-user.target
`,
			Name:    "etc-kubernetes-data-etcd.automount",
			Enable:  true,
			Command: "start",
		},
		// Mount for etcd volume activated by automount
		{
			AssetContent: `[Unit]
Description=Mount for etcd volume
[Mount]
What=etcdshare
Where=/etc/kubernetes/data/etcd
Options=trans=virtio,version=9p2000.L,cache=mmap
Type=9p
[Install]
WantedBy=multi-user.target
`,
			Name:   "etc-kubernetes-data-etcd.mount",
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
	/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system delete pod -l k8s-app=kube-proxy"

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

func (e *v2_0_0masterExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	return nil
}
