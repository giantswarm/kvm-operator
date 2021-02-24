package key

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	RequeueNone = reconcile.Result{
		Requeue:      false,
		RequeueAfter: 0,
	}
	RequeueShort = reconcile.Result{
		Requeue:      true,
		RequeueAfter: time.Second * 10,
	}
	RequeueLong = reconcile.Result{
		Requeue:      true,
		RequeueAfter: time.Second * 30,
	}
)
