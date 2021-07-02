package cloudconfigtest

import (
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v10/pkg/template"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/kvm-operator/v4/service/controller/cloudconfig"
)

func New() *cloudconfig.CloudConfig {
	c := cloudconfig.DefaultConfig()

	c.Logger = microloggertest.New()
	c.DockerhubToken = "token"

	packagePath, err := k8scloudconfig.GetPackagePath()
	if err != nil {
		panic(err)
	}
	c.IgnitionPath = packagePath

	newCloudConfig, err := cloudconfig.New(c)
	if err != nil {
		panic(err)
	}

	return newCloudConfig
}
