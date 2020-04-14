package cloudconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_4_9_2"
	"github.com/giantswarm/randomkeys"
)

type IgnitionTemplateData struct {
	CustomObject v1alpha1.KVMConfig
	ClusterCerts certs.Cluster
	ClusterKeys  randomkeys.Cluster
	Images       k8scloudconfig.Images
}
