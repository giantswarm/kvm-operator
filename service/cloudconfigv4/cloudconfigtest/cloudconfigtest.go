package cloudconfigtest

import (
	"github.com/giantswarm/kvm-operator/service/cloudconfigv4"
	"github.com/giantswarm/micrologger/microloggertest"
)

func New() *cloudconfigv4.CloudConfig {
	c := cloudconfigv4.DefaultConfig()

	c.Logger = microloggertest.New()

	newCloudConfig, err := cloudconfigv4.New(c)
	if err != nil {
		panic(err)
	}

	return newCloudConfig
}
