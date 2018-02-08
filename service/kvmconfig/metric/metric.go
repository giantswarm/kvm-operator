package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	prometheusNamespace = "kvm_operator"
	prometheusSubsystem = "deployment_resource"
)

var VersionBundleVersionGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Subsystem: prometheusSubsystem,
		Name:      "version_bundle_version_total",
		Help:      "A metric labeled by major, minor and patch version of the version bundle being in use.",
	},
	[]string{"major", "minor", "patch"},
)

func init() {
	prometheus.MustRegister(VersionBundleVersionGauge)
}
