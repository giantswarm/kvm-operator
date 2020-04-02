package cloudconfig

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_4_9_1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/service/controller/key"
)

// NewMasterTemplate generates a new worker cloud config template and returns it
// as a base64 encoded string.
func (c *CloudConfig) NewMasterTemplate(data IgnitionTemplateData, node v1alpha1.ClusterNode, nodeIndex int) (string, error) {
	var err error

	var params k8scloudconfig.Params
	{
		params = k8scloudconfig.DefaultParams()

		params.APIServerEncryptionKey = string(data.ClusterKeys.APIServerEncryptionKey)
		params.BaseDomain = key.BaseDomain(data.CustomObject)
		params.Cluster = data.CustomObject.Spec.Cluster
		// Ingress controller service remains in k8scloudconfig and will be
		// removed in a later migration.
		params.DisableIngressControllerService = false
		params.Extension = &masterExtension{
			certs:        data.ClusterCerts,
			customObject: data.CustomObject,
			nodeIndex:    nodeIndex,
		}
		params.Node = node
		params.Hyperkube.Apiserver.Pod.CommandExtraArgs = c.k8sAPIExtraArgs
		params.SSOPublicKey = c.ssoPublicKey

		ignitionPath := k8scloudconfig.GetIgnitionPath(c.ignitionPath)
		params.Files, err = k8scloudconfig.RenderFiles(ignitionPath, params)
		if err != nil {
			return "", microerror.Mask(err)
		}
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
	certs        certs.Cluster
	customObject v1alpha1.KVMConfig
	nodeIndex    int
}

func (e *masterExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	var filesMeta []k8scloudconfig.FileMetadata

	for _, f := range certs.NewFilesClusterMaster(e.certs) {
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
		AssetContent: fmt.Sprintf("InitiatorName=%s", key.IscsiInitiatorName(e.customObject, e.nodeIndex, key.MasterID)),
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
Environment="KUBECONFIG=/etc/kubernetes/config/addons-kubeconfig.yml"
Environment="KUBECTL=quay.io/giantswarm/docker-kubectl:e777d4eaf369d4dabc393c5da42121c2a725ea6a"
ExecStart=/bin/sh -c '\
	while [ "$(/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL get cs | grep Healthy | wc -l)" -ne "3" ]; do sleep 1 && echo "Waiting for healthy k8s";done;sleep 30s; \
	RETRY=5;result="";\
	while [ "$result" != "ok" ] && [ $RETRY -gt 0 ]; do\
		sleep 10s; echo "Trying to restart k8s services ...";\
		let RETRY=$RETRY-1;\
		/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system delete pod -l k8s-app=calico-node && \
		sleep 1m && \
		/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system delete pod -l k8s-app=kube-proxy && \
		/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system delete pod -l k8s-app=calico-kube-controllers && \
		/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system delete pod -l k8s-app=coredns &&\
		result="ok" || echo "failed";\
	done;\
	[ "$result" != "ok" ] && echo "Failed to restart k8s services." && exit 1 || echo "Successfully restarted k8s services.";'

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
