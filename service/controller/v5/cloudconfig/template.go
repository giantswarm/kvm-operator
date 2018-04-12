package cloudconfig

import (
	"bytes"
	"text/template"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
)

const (
	set_ownership_etcd_data_dir_dropin = `[Unit]
Requires=etc-kubernetes-data-etcd.mount
After=etc-kubernetes-data-etcd.mount
`
	encryptionConfigTemplate = `
kind: EncryptionConfig
apiVersion: v1
resources:
  - resources:
    - secrets
    providers:
    - aescbc:
        keys:
        - name: key1
          secret: {{.EncryptionKey}}
    - identity: {}
`
)

func encryptionConfig(encryptionKey randomkeys.RandomKey) (string, error) {
	tmpl, err := template.New("encryptionConfig").Parse(encryptionConfigTemplate)
	if err != nil {
		return "", microerror.Mask(err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, struct {
		EncryptionKey string
	}{
		EncryptionKey: string(encryptionKey),
	})
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(buf.Bytes()), nil
}
