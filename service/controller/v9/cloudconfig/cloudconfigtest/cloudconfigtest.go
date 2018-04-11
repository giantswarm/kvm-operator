package cloudconfigtest

import (
	"github.com/giantswarm/kvm-operator/service/kvmconfig/v9/cloudconfig"
	"github.com/giantswarm/micrologger/microloggertest"
)

func New() *cloudconfig.CloudConfig {
	c := cloudconfig.DefaultConfig()

	c.Logger = microloggertest.New()

	newCloudConfig, err := cloudconfig.New(c)
	if err != nil {
		panic(err)
	}

	return newCloudConfig
}
