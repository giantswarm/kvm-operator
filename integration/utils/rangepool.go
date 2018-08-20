package utils

import (
	"context"
	"fmt"
	"net"

	"k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/crdstorage"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/rangepool"
)

const (
	vniMin      = 1
	vniMax      = 1000
	nodePortMin = 30100
	nodePortMax = 31500

	flannelCidrSize        = 26
	flannelE2eNetworkRange = "10.1.0.0/16"

	gsNamespace = "giantswarm"
)

func InitCRDStorage(h *framework.Host, l micrologger.Logger) (microstorage.Storage, error) {
	var err error

	k8sExtClient, err := apiextensionsclient.NewForConfig(h.RestConfig())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var k8sCrdClient *k8scrdclient.CRDClient
	{
		var k8sCrdClientConfig k8scrdclient.Config
		k8sCrdClientConfig.Logger = l
		k8sCrdClientConfig.K8sExtClient = k8sExtClient

		k8sCrdClient, err = k8scrdclient.New(k8sCrdClientConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := crdstorage.DefaultConfig()
	c.CRDClient = k8sCrdClient
	c.G8sClient = h.G8sClient()
	c.K8sClient = h.K8sClient()
	c.Logger = l

	c.Name = "kvm-e2e"
	c.Namespace = &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: gsNamespace,
		},
	}

	crdStorage, err := crdstorage.New(c)

	if err != nil {
		return nil, microerror.Mask(err)
	}

	l.Log("info", "booting crdstorage")
	err = crdStorage.Boot(context.Background())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return crdStorage, nil
}

func InitRangePool(crdStorage microstorage.Storage, l micrologger.Logger) (*rangepool.Service, error) {

	rangePoolConfig := rangepool.DefaultConfig()
	rangePoolConfig.Logger = l
	rangePoolConfig.Storage = crdStorage

	rangePool, err := rangepool.New(rangePoolConfig)
	if err != nil {
		return nil, microerror.Mask(err)

	}

	return rangePool, nil
}

func GenerateVNI(rangePool *rangepool.Service, clusterID string) (int, error) {
	items, err := rangePool.Create(
		context.Background(),
		rangePoolVNIID(clusterID),
		rangePoolVNIID(clusterID),
		1, // num
		vniMin,
		vniMax,
	)

	if err != nil {
		return 0, microerror.Mask(err)
	}

	if len(items) != 1 {
		return 0, microerror.Maskf(executionFailedError, "range pool VNI generation failed, expected 1 got %d", len(items))
	}

	return items[0], nil
}

func DeleteVNI(rangePool *rangepool.Service, clusterID string) error {
	return rangePool.Delete(context.Background(), gsNamespace, rangePoolVNIID(clusterID))
}

func GenerateIngressNodePorts(rangePool *rangepool.Service, clusterID string) (int, int, error) {
	items, err := rangePool.Create(
		context.Background(),
		gsNamespace,
		rangePoolIngressID(clusterID),
		2, // num
		nodePortMin,
		nodePortMax,
	)
	if err != nil {
		return 0, 0, microerror.Mask(err)
	}

	if len(items) != 2 {
		return 0, 0, microerror.Maskf(executionFailedError, "range pool ingress port generation failed, expected 2 got %d", len(items))
	}

	return items[0], items[1], nil
}

func DeleteIngressNodePorts(rangePool *rangepool.Service, clusterID string) error {
	return rangePool.Delete(context.Background(), gsNamespace, rangePoolIngressID(clusterID))
}

func GenerateFlannelNetwork(clusterID string, crdStorage microstorage.Storage, l micrologger.Logger) (string, error) {
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

	cidr, err := ipamService.CreateSubnet(context.Background(), cidrMask, flannelNetworkAnnotation(clusterID))
	if err != nil {
		return "", microerror.Mask(err)
	}

	return cidr.String(), nil
}

func DeleteFlannelNetwork(network string, crdStorage microstorage.Storage, l micrologger.Logger) error {
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

	err = ipamService.DeleteSubnet(context.Background(), *subnet)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func rangePoolVNIID(clusterID string) string {
	return fmt.Sprintf("%s-vni", clusterID)
}
func rangePoolIngressID(clusterID string) string {
	return fmt.Sprintf("%s-ingress", clusterID)
}
func flannelNetworkAnnotation(clusterID string) string {
	return fmt.Sprintf("kvm-e2e-%s", clusterID)
}
