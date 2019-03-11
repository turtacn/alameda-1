package metadata

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

var (
	alamedaRecommendationGVR = schema.GroupVersionResource{}
	resourcesKindMapMutex    = &sync.Mutex{}
)

// OwnerReferenceTracer struct to trace owner references
type OwnerReferenceTracer struct {
	k8sClient          kubernetes.Interface
	k8sDynamicClient   dynamic.Interface
	k8sDiscoveryClient *discovery.DiscoveryClient
	k8sRestMapper      meta.RESTMapper
}

// NewDefaultOwnerReferenceTracer build OwnerReferenceTracer
func NewDefaultOwnerReferenceTracer() (*OwnerReferenceTracer, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Errorf("new OwnerReferenceTracer failed: %s", err.Error())
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Errorf("new OwnerReferenceTracer failed: %s", err.Error())
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.Errorf("new OwnerReferenceTracer failed: %s", err.Error())
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, errors.Errorf("new OwnerReferenceTracer failed: %s", err.Error())
	}

	gr, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, errors.Errorf("new OwnerReferenceTracer failed: %s", err.Error())
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(gr)

	o := &OwnerReferenceTracer{
		k8sClient:          client,
		k8sDynamicClient:   dynamicClient,
		k8sDiscoveryClient: discoveryClient,
		k8sRestMapper:      restMapper,
	}

	return o, nil
}

// NewOwnerReferenceTracerWithConfig build OwnerReferenceTracer
func NewOwnerReferenceTracerWithConfig(cfg rest.Config) (*OwnerReferenceTracer, error) {

	copyCfg := cfg

	client, err := kubernetes.NewForConfig(&copyCfg)
	if err != nil {
		return nil, errors.Errorf("new resource recommendator failed: %s", err.Error())
	}

	dynamicClient, err := dynamic.NewForConfig(&copyCfg)
	if err != nil {
		return nil, errors.Errorf("new resource recommendator failed: %s", err.Error())
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(&copyCfg)
	if err != nil {
		return nil, errors.Errorf("new resource recommendator failed: %s", err.Error())
	}

	gr, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, errors.Errorf("new OwnerReferenceTracer failed: %s", err.Error())
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(gr)

	impl := &OwnerReferenceTracer{
		k8sClient:          client,
		k8sDynamicClient:   dynamicClient,
		k8sDiscoveryClient: discoveryClient,
		k8sRestMapper:      restMapper,
	}

	return impl, nil
}

// GetRootControllerKindAndNameOfOwnerReferences gets root owner references that is Controller
func (ort *OwnerReferenceTracer) GetRootControllerKindAndNameOfOwnerReferences(namespace string, ownerRefs []meta_v1.OwnerReference) (kind, name string, err error) {

	var controllerOwnerRef *meta_v1.OwnerReference
	finish := false
	for !finish {

		if len(ownerRefs) == 0 {
			finish = true
			break
		}

		// get owner that is controller
		for _, ownerRef := range ownerRefs {
			if ownerRef.Controller != nil && *ownerRef.Controller {
				controllerOwnerRef = &ownerRef
				break
			}
		}

		// there is no ownerReference that is Controller, need no tracing
		if controllerOwnerRef == nil {
			finish = true
			break
		}

		gvk := schema.FromAPIVersionAndKind(controllerOwnerRef.APIVersion, controllerOwnerRef.Kind)
		ownerRefs, err = ort.getOwnerRefsOfResource(namespace, controllerOwnerRef.Name, gvk)
		if err != nil {
			return "", "", errors.Wrap(err, "get root controller name from owner references failed")
		}
	}

	if controllerOwnerRef != nil {
		kind = controllerOwnerRef.Kind
		name = controllerOwnerRef.Name
	}

	return kind, name, err
}

func (ort *OwnerReferenceTracer) getOwnerRefsOfResource(namespace, name string, gvk schema.GroupVersionKind) ([]meta_v1.OwnerReference, error) {

	ownerRefs := make([]meta_v1.OwnerReference, 0)

	restMapping, err := ort.k8sRestMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return ownerRefs, errors.Errorf("get owner references of %s/%s gvk: %s failed: %s", namespace, name, gvk.String(), err.Error())
	}

	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: fmt.Sprintf("namespaces/%s/%s", namespace, restMapping.Resource.Resource),
	}
	us, err := ort.k8sDynamicClient.Resource(gvr).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return ownerRefs, errors.Errorf("get owner references of resource %s in namespace %s failed: %s", gvr.String(), namespace, err.Error())
	}
	ownerRefs = us.GetOwnerReferences()

	return ownerRefs, nil
}
