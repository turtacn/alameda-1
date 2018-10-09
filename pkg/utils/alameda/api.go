package alameda

import (
	"time"

	as_types "github.com/containers-ai/alameda/pkg/apis/autoscaling/v1alpha1"
	as_clientset "github.com/containers-ai/alameda/pkg/client/clientset/versioned"
	as_lister "github.com/containers-ai/alameda/pkg/client/listers/autoscaling/v1alpha1"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/apis/core"
)

func NewAllVpasLister(asClientSet *as_clientset.Clientset, stopChannel <-chan struct{}) as_lister.AlamedaVPALister {
	vpaListWatch := cache.NewListWatchFromClient(asClientSet.AutoscalingV1alpha1().RESTClient(), "alamedavpas", core.NamespaceAll, fields.Everything())
	indexer, controller := cache.NewIndexerInformer(vpaListWatch,
		&as_types.AlamedaVPA{},
		1*time.Hour,
		&cache.ResourceEventHandlerFuncs{
			AddFunc: handleObject,
			UpdateFunc: func(old, new interface{}) {
				handleObject(new)
			},
			DeleteFunc: handleObject,
		},
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	vpaLister := as_lister.NewAlamedaVPALister(indexer)
	go controller.Run(stopChannel)
	if !cache.WaitForCacheSync(make(chan struct{}), controller.HasSynced) {
		glog.Fatalf("Failed to sync alamedavpa cache during initialization")
	} else {
		glog.Info("Initial alamedavpa synced successfully")
	}
	return vpaLister
}

func handleObject(obj interface{}) {
	object := obj.(metav1.Object)
	glog.V(4).Infof("Processing object: %s", object.GetName())
}
