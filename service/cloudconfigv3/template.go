package cloudconfigv3

const (
	set_ownership_etcd_data_dir_dropin = `[Unit]
Requires=etc-kubernetes-data-etcd.mount
After=etc-kubernetes-data-etcd.mount
`
)
