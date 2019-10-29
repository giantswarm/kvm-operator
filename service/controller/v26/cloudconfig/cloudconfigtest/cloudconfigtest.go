package cloudconfigtest

import (
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_4_9_0"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/kvm-operator/service/controller/v26/cloudconfig"
)

func New() *cloudconfig.CloudConfig {
	c := cloudconfig.DefaultConfig()

	c.Logger = microloggertest.New()

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
