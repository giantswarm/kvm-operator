package ipam

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
)

const (
	flannelCidrSize        = 26
	flannelE2eNetworkRange = "10.1.0.0/16"
)

func GenerateFlannelNetwork(ctx context.Context, clusterID string, crdStorage microstorage.Storage, l micrologger.Logger) (string, error) {
	var err error
	var ipamConfig ipam.Config
	{
		ipamConfig.Logger = l
		ipamConfig.Storage = crdStorage

		var network *net.IPNet
		_, network, err = net.ParseCIDR(flannelE2eNetworkRange)
		if err != nil {
			return "", microerror.Mask(err)
		}
		ipamConfig.Network = network
	}
	ipamService, err := ipam.New(ipamConfig)
	if err != nil {
		return "", microerror.Mask(err)
	}

	cidrMask := net.CIDRMask(flannelCidrSize, 32)

	var cidr net.IPNet
	o := func() error {
		var err error
		cidr, err = ipamService.CreateSubnet(ctx, cidrMask, flannelNetworkAnnotation(clusterID))
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}

	b := backoff.NewExponential(framework.ShortMaxWait, framework.ShortMaxInterval)
	n := backoff.NewNotifier(l, context.Background())
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return cidr.String(), nil
}

func DeleteFlannelNetwork(ctx context.Context, network string, crdStorage microstorage.Storage, l micrologger.Logger) error {
	var err error
	var ipamConfig ipam.Config
	{
		ipamConfig.Logger = l
		ipamConfig.Storage = crdStorage

		var network *net.IPNet
		_, network, err = net.ParseCIDR(flannelE2eNetworkRange)
		if err != nil {
			microerror.Mask(err)
		}
		ipamConfig.Network = network
	}
	ipamService, err := ipam.New(ipamConfig)
	if err != nil {
		return microerror.Mask(err)
	}
	_, subnet, err := net.ParseCIDR(network)
	if err != nil {
		return microerror.Mask(err)
	}

	o := func() error {
		err = ipamService.DeleteSubnet(ctx, *subnet)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}

	b := backoff.NewExponential(framework.ShortMaxWait, framework.ShortMaxInterval)
	n := backoff.NewNotifier(l, context.Background())
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func flannelNetworkAnnotation(clusterID string) string {
	return fmt.Sprintf("kvm-e2e-%s", clusterID)
}
