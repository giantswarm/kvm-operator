package cloudconfigtest

import (
	"github.com/giantswarm/kvm-operator/service/cloudconfigv3"
	"github.com/giantswarm/micrologger/microloggertest"
)

func New() *cloudconfigv3.CloudConfig {
	c := cloudconfigv3.DefaultConfig()

	c.Logger = microloggertest.New()

	newCloudConfig, err := cloudconfigv3.New(c)
	if err != nil {
		panic(err)
	}

	return newCloudConfig
}
