package cloudconfig

import (
	"context"
	"sync"

	"github.com/giantswarm/certs/v3/pkg/certs"
	"github.com/giantswarm/microerror"
	"golang.org/x/sync/errgroup"
)

type certFileMapping map[certs.Cert]func(certs.TLS) []certs.File

var masterCertFiles = certFileMapping{
	certs.APICert:              certs.NewFilesAPI,
	certs.EtcdCert:             certs.NewFilesEtcd,
	certs.ServiceAccountCert:   certs.NewFilesServiceAccount,
	certs.CalicoEtcdClientCert: certs.NewFilesCalicoEtcdClient,
}

var workerCertFiles = certFileMapping{
	certs.WorkerCert:           certs.NewFilesWorker,
	certs.CalicoEtcdClientCert: certs.NewFilesCalicoEtcdClient,
}

func fetchCertFiles(ctx context.Context, searcher certs.Interface, clusterID string, mapping certFileMapping) ([]certs.File, error) {
	group, groupCtx := errgroup.WithContext(ctx)

	var mu sync.Mutex
	var certFiles []certs.File
	for cert, fileMapper := range mapping {
		cert, fileMapper := cert, fileMapper
		group.Go(func() error {
			tls, err := searcher.SearchTLS(groupCtx, clusterID, cert)
			if err != nil {
				return microerror.Mask(err)
			}

			mu.Lock()
			certFiles = append(certFiles, fileMapper(tls)...)
			mu.Unlock()

			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return certFiles, nil
}
