package cloudconfigv1

const (
	etcd_data_dir_dropin = `[Unit]
Before=set-ownership-etcd-data-dir.service
`
	set_ownership_etcd_data_dir_dropin = `[Unit]
Requires=etc-kubernetes-data-etcd.mount
After=etc_kubernetes_data_etcd.mount
`
)
