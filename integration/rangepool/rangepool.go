package rangepool

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/rangepool"
)

const (
	vniMin      = 1
	vniMax      = 1000
	nodePortMin = 30100
	nodePortMax = 31500

	vniRangepoolNamespace     = "vni"
	ingressRangepoolNamespace = "ingress"
)

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

func GenerateVNI(ctx context.Context, rangePool *rangepool.Service, clusterID string) (int, error) {
	items, err := rangePool.Create(
		ctx,
		vniRangepoolNamespace,
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

func DeleteVNI(ctx context.Context, rangePool *rangepool.Service, clusterID string) error {
	err := rangePool.Delete(ctx, vniRangepoolNamespace, rangePoolVNIID(clusterID))
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func GenerateIngressNodePorts(ctx context.Context, rangePool *rangepool.Service, clusterID string) (int, int, error) {
	items, err := rangePool.Create(
		ctx,
		ingressRangepoolNamespace,
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

func DeleteIngressNodePorts(ctx context.Context, rangePool *rangepool.Service, clusterID string) error {
	err := rangePool.Delete(ctx, ingressRangepoolNamespace, rangePoolIngressID(clusterID))
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
