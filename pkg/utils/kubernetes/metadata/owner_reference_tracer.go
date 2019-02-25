package metadata

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	// apiGroupVersion_Resource_KindMap write once in method initResourcesKindMap
	apiGroupVersion_Resource_KindMap map[string]map[string]string
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

	o := &OwnerReferenceTracer{
		k8sClient:                        client,
		k8sDynamicClient:                 dynamicClient,
		k8sDiscoveryClient:               discoveryClient,
		apiGroupVersion_Resource_KindMap: make(map[string]map[string]string),
	}

	o.initResourcesKindMap()

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

	impl := &OwnerReferenceTracer{
		k8sClient:          client,
		k8sDynamicClient:   dynamicClient,
		k8sDiscoveryClient: discoveryClient,
	}

	impl.initResourcesKindMap()

	return impl, nil
}

func (ort *OwnerReferenceTracer) initResourcesKindMap() error {

	apiResourceLists, err := ort.k8sDiscoveryClient.ServerResources()
	if err != nil {
		return errors.Errorf("initialize Kubernetes resource kind mapping failed: %s", err.Error())
	}

	resourcesKindMapMutex.Lock()
	defer resourcesKindMapMutex.Unlock()

	for _, apiResourceList := range apiResourceLists {
		gv := apiResourceList.GroupVersion
		if _, exist := ort.apiGroupVersion_Resource_KindMap[gv]; !exist {
			ort.apiGroupVersion_Resource_KindMap[gv] = make(map[string]string)
		}
		for _, resource := range apiResourceList.APIResources {
			ort.apiGroupVersion_Resource_KindMap[gv][resource.Name] = resource.Kind
		}
	}

	return nil
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

	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: fmt.Sprintf("namespaces/%s/%s", namespace, ort.findPossibleResourceOfGVKInLocalCache(gvk)),
	}
	us, err := ort.k8sDynamicClient.Resource(gvr).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return ownerRefs, errors.Errorf("get owner references of resource %s in namespace %s failed: %s", gvr.String(), namespace, err.Error())
	}
	ownerRefs = us.GetOwnerReferences()

	return ownerRefs, nil
}

func (ort *OwnerReferenceTracer) findPossibleResourceOfGVKInLocalCache(gvk schema.GroupVersionKind) string {

	candidatesResource := make([]string, 0)

	if resourceKindMap, exist := ort.apiGroupVersion_Resource_KindMap[gvk.GroupVersion().String()]; exist {
		for resourceName, kindName := range resourceKindMap {
			if kindName == gvk.Kind {
				candidatesResource = append(candidatesResource, resourceName)
			}
		}
	}

	for _, resource := range candidatesResource {
		if !strings.Contains(resource, "/") {
			return resource
		}
	}

	return ""
}
