package api

import (
	"github.com/giantswarm/kvm-operator/flag/service/installation/guest/kubernetes/api/auth"
)

type API struct {
	Auth auth.Auth
}
