package server

import (
	openshift_apps "github.com/openshift/api/apps"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

func init() {
	addToScheme(scheme)
}

func addToScheme(scheme *runtime.Scheme) {
	core_v1.AddToScheme(scheme)
	apps_v1.AddToScheme(scheme)
	admissionregistrationv1beta1.AddToScheme(scheme)
	err := openshift_apps.Install(scheme)
	if err != nil {
		panic(err)
	}
}
