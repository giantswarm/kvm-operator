package cloudconfig

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/giantswarm/microerror"
)

const (
	etcd_data_dir_dropin = `[Unit]
Before=set-ownership-etcd-data-dir.service
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

func EncryptionConfig(encryptionKey string) (string, error) {
	fmt.Println("jgsqware-1", encryptionKey)
	tmpl, err := template.New("encryptionConfig").Parse(encryptionConfigTemplate)
	if err != nil {
		return "", microerror.Mask(err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, struct {
		EncryptionKey string
	}{
		EncryptionKey: encryptionKey,
	})
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(buf.Bytes()), nil
}
