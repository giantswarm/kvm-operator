package cloudconfigv2

const (
	etcd_data_dir_dropin = `[Unit]
Before=set_ownership_etcd_data_dir.service
`
)
