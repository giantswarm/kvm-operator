package cloudconfig

import (
	"github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs/v3/pkg/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v10/pkg/template"
	"github.com/giantswarm/randomkeys/v2"
)

type IgnitionTemplateData struct {
	CustomResource v1alpha2.KVMMachine
	CertsSearcher  certs.Interface
	ClusterKeys    randomkeys.Cluster
	Images         k8scloudconfig.Images
	Versions       k8scloudconfig.Versions
}
