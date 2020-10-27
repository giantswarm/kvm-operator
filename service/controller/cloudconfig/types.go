package cloudconfig

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/v3/pkg/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v9/pkg/template"
	"github.com/giantswarm/randomkeys/v2"
)

type IgnitionTemplateData struct {
	CustomObject  v1alpha1.KVMConfig
	CertsSearcher certs.Interface
	ClusterKeys   randomkeys.Cluster
	Images        k8scloudconfig.Images
	Versions      k8scloudconfig.Versions
}
