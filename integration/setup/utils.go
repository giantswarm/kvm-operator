package setup

import (
	"fmt"
	"path"

	"github.com/spf13/afero"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/kvm-operator/integration/env"
)

const kubeConfigTmpl string = `
apiVersion: v1
kind: Config
clusters:
- name: giantswarm-e2e-kvm
  cluster:
    server: %s
    certificate-authority-data: %s
users:
- name: giantswarm-e2e-kvm-user
  user:
    client-certificate-data: %s
    client-key-data: %s
contexts:
- name: giantswarm-e2e-kvm
  context:
    cluster: giantswarm-e2e-kvm
    user: giantswarm-e2e-kvm-user
current-context: giantswarm-e2e-kvm
preferences: {}
`

// create kubeconfig from env values
func createKubeconfig(filePath string) error {
	// fill template with values
	kubeConfigContet := fmt.Sprintf(kubeConfigTmpl, env.K8sApiUrl(), env.K8sCertCa(), env.K8sCert(), env.K8sCertPrivate())

	var aferoFs = afero.NewOsFs()

	f, err := aferoFs.Create(filePath)
	if err != nil {
		return microerror.Maskf(err, fmt.Sprintf("Failed to create kubeconfig %s", filePath))
	}

	_, err = f.WriteString(kubeConfigContet)
	if err != nil {
		return microerror.Maskf(err, fmt.Sprintf("Failed to write content of kubeconfig %s", filePath))
	}

	return nil
}

func kubeconfigFilePath(baseDir string) string {
	// values copied from here https://github.com/giantswarm/e2e-harness/blob/master/pkg/cluster/cluster.go#L183
	return path.Join(baseDir, "workdir", ".shipyard", "config")
}
