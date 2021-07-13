package api

import (
	"github.com/giantswarm/kvm-operator/v4/flag/service/installation/workload/kubernetes/api/auth"
)

type API struct {
	Auth auth.Auth
}
