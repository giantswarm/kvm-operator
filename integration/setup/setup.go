// +build k8srequired

package setup

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	gotemplate "text/template"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/crdstorage"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/operatorkit/client/k8scrdclient"
	"github.com/giantswarm/rangepool"
	"k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/kvm-operator/integration/env"
	"github.com/giantswarm/kvm-operator/integration/teardown"
	"github.com/giantswarm/kvm-operator/integration/template"
)

const (
	kvmResourceValuesFile = "/tmp/kvm-operator-values.yaml"

	vniMin      = 1
	vniMax      = 1000
	nodePortMin = 30100
	nodePortMax = 31500
)

// WrapTestMain setup and teardown e2e testing environment.
func WrapTestMain(g *framework.Guest, h *framework.Host, m *testing.M) {
	var r int

	err := Setup(g, h)
	if err != nil {
		log.Printf("%#v\n", err)
		r = 1
	} else {
		r = m.Run()
	}

	if env.KeepResources() != "true" {
		teardown.Teardown(g, h)
	}

	os.Exit(r)
}

// Setup e2e testing environment.
func Setup(g *framework.Guest, h *framework.Host) error {
	var err error

	err = Resources(g, h)
	if err != nil {
		return microerror.Mask(err)
	}

	err = g.Setup()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// Resources install required charts.
func Resources(g *framework.Guest, h *framework.Host) error {
	var err error

	{
		err = h.InstallStableOperator("cert-operator", "certconfig", e2etemplates.CertOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}

		err = h.InstallBranchOperator("kvm-operator", "kvmconfig", template.KVMOperatorChartValues)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = h.InstallCertResource()
		if err != nil {
			return microerror.Mask(err)
		}

		err = installKVMResource(h)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func installKVMResource(h *framework.Host) error {
	var err error

	var l micrologger.Logger
	{
		c := micrologger.Config{}

		l, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var kvmResourceChartValues template.KVMConfigE2eChartValues
	{
		kvmResourceChartValues.ClusterID = env.ClusterID()

		rangePool, err := initRangePool(h, l)
		if err != nil {
			return microerror.Mask(err)
		}

		{
			vni, err := generateVNI(rangePool, env.ClusterID())
			if err != nil {
				return microerror.Mask(err)
			}
			kvmResourceChartValues.VNI = vni
		}

		{
			httpPort, httpsPort, err := generateIngressNodePorts(rangePool, env.ClusterID())
			if err != nil {
				return microerror.Mask(err)
			}
			kvmResourceChartValues.HttpNodePort = httpPort
			kvmResourceChartValues.HttpsNodePort = httpsPort
		}
	}

	o := func() error {
		// NOTE we ignore errors here because we cannot get really useful error
		// handling done. This here should anyway only be a quick fix until we use
		// the helm client lib. Then error handling will be better.
		framework.HelmCmd(fmt.Sprintf("delete --purge %s-kvm-config-e2e", h.TargetNamespace()))

		var buffer *bytes.Buffer

		tmpl := gotemplate.New("kvm-e2e-values")
		err := tmpl.Execute(buffer, kvmResourceChartValues)
		if err != nil {
			return microerror.Mask(err)
		}

		err = ioutil.WriteFile(kvmResourceValuesFile, buffer.Bytes(), 0644)
		if err != nil {
			return microerror.Mask(err)
		}

		err = framework.HelmCmd(fmt.Sprintf("registry install quay.io/giantswarm/apiextensions-kvm-config-e2e-chart:stable -- -n %s-kvm-config-e2e --values %s", h.TargetNamespace(), kvmResourceValuesFile))
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

func initRangePool(h *framework.Host, l micrologger.Logger) (*rangepool.Service, error) {
	var err error
	var storage microstorage.Storage
	{
		var c crdstorage.Config

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

		c.CRDClient = k8sCrdClient
		c.G8sClient = h.G8sClient()
		c.K8sClient = h.K8sClient()
		c.Logger = l

		c.Name = "kvm-e2e"
		c.Namespace = &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "giantswarm",
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

		storage = crdStorage
	}
	var rangePool *rangepool.Service
	{
		rangePoolConfig := rangepool.DefaultConfig()
		rangePoolConfig.Logger = l
		rangePoolConfig.Storage = storage

		rangePool, err = rangepool.New(rangePoolConfig)
		if err != nil {
			return nil, microerror.Mask(err)

		}
	}

	return rangePool, nil
}

func generateVNI(rangePool *rangepool.Service, clusterID string) (int, error) {
	items, err := rangePool.Create(
		context.Background(),
		clusterID,
		clusterID,
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

func generateIngressNodePorts(rangePool *rangepool.Service, clusterID string) (int, int, error) {
	items, err := rangePool.Create(
		context.Background(),
		clusterID,
		clusterID,
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
