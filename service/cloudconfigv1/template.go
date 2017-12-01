package cloudconfig

const (
	etcd_data_dir_dropin = `[Unit]
Before=set-ownership-etcd-data-dir.service
`
)
