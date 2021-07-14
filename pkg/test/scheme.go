package test

import (
	giantswarmscheme "github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned/scheme"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
)

var Scheme = runtime.NewScheme()
var localSchemeBuilder = runtime.SchemeBuilder{
	k8sscheme.AddToScheme,        // adds all kubernetes GVKs (Pod, Node, etc.)
	giantswarmscheme.AddToScheme, // adds all giantswarm GVKs (Release, App, etc.)
}

var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	err := AddToScheme(Scheme)
	if err != nil {
		panic(microerror.Mask(err))
	}
}
