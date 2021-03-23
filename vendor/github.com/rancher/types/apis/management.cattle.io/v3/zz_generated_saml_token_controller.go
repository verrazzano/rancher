package v3

import (
	"context"
	"time"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"github.com/rancher/norman/resource"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	SamlTokenGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "SamlToken",
	}
	SamlTokenResource = metav1.APIResource{
		Name:         "samltokens",
		SingularName: "samltoken",
		Namespaced:   true,

		Kind: SamlTokenGroupVersionKind.Kind,
	}

	SamlTokenGroupVersionResource = schema.GroupVersionResource{
		Group:    GroupName,
		Version:  Version,
		Resource: "samltokens",
	}
)

func init() {
	resource.Put(SamlTokenGroupVersionResource)
}

func NewSamlToken(namespace, name string, obj SamlToken) *SamlToken {
	obj.APIVersion, obj.Kind = SamlTokenGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type SamlTokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SamlToken `json:"items"`
}

type SamlTokenHandlerFunc func(key string, obj *SamlToken) (runtime.Object, error)

type SamlTokenChangeHandlerFunc func(obj *SamlToken) (runtime.Object, error)

type SamlTokenLister interface {
	List(namespace string, selector labels.Selector) (ret []*SamlToken, err error)
	Get(namespace, name string) (*SamlToken, error)
}

type SamlTokenController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() SamlTokenLister
	AddHandler(ctx context.Context, name string, handler SamlTokenHandlerFunc)
	AddFeatureHandler(ctx context.Context, enabled func() bool, name string, sync SamlTokenHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler SamlTokenHandlerFunc)
	AddClusterScopedFeatureHandler(ctx context.Context, enabled func() bool, name, clusterName string, handler SamlTokenHandlerFunc)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, after time.Duration)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type SamlTokenInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*SamlToken) (*SamlToken, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*SamlToken, error)
	Get(name string, opts metav1.GetOptions) (*SamlToken, error)
	Update(*SamlToken) (*SamlToken, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*SamlTokenList, error)
	ListNamespaced(namespace string, opts metav1.ListOptions) (*SamlTokenList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() SamlTokenController
	AddHandler(ctx context.Context, name string, sync SamlTokenHandlerFunc)
	AddFeatureHandler(ctx context.Context, enabled func() bool, name string, sync SamlTokenHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle SamlTokenLifecycle)
	AddFeatureLifecycle(ctx context.Context, enabled func() bool, name string, lifecycle SamlTokenLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync SamlTokenHandlerFunc)
	AddClusterScopedFeatureHandler(ctx context.Context, enabled func() bool, name, clusterName string, sync SamlTokenHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle SamlTokenLifecycle)
	AddClusterScopedFeatureLifecycle(ctx context.Context, enabled func() bool, name, clusterName string, lifecycle SamlTokenLifecycle)
}

type samlTokenLister struct {
	controller *samlTokenController
}

func (l *samlTokenLister) List(namespace string, selector labels.Selector) (ret []*SamlToken, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*SamlToken))
	})
	return
}

func (l *samlTokenLister) Get(namespace, name string) (*SamlToken, error) {
	var key string
	if namespace != "" {
		key = namespace + "/" + name
	} else {
		key = name
	}
	obj, exists, err := l.controller.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group:    SamlTokenGroupVersionKind.Group,
			Resource: "samlToken",
		}, key)
	}
	return obj.(*SamlToken), nil
}

type samlTokenController struct {
	controller.GenericController
}

func (c *samlTokenController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *samlTokenController) Lister() SamlTokenLister {
	return &samlTokenLister{
		controller: c,
	}
}

func (c *samlTokenController) AddHandler(ctx context.Context, name string, handler SamlTokenHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*SamlToken); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *samlTokenController) AddFeatureHandler(ctx context.Context, enabled func() bool, name string, handler SamlTokenHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if !enabled() {
			return nil, nil
		} else if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*SamlToken); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *samlTokenController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler SamlTokenHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*SamlToken); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *samlTokenController) AddClusterScopedFeatureHandler(ctx context.Context, enabled func() bool, name, cluster string, handler SamlTokenHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if !enabled() {
			return nil, nil
		} else if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*SamlToken); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type samlTokenFactory struct {
}

func (c samlTokenFactory) Object() runtime.Object {
	return &SamlToken{}
}

func (c samlTokenFactory) List() runtime.Object {
	return &SamlTokenList{}
}

func (s *samlTokenClient) Controller() SamlTokenController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.samlTokenControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(SamlTokenGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &samlTokenController{
		GenericController: genericController,
	}

	s.client.samlTokenControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type samlTokenClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   SamlTokenController
}

func (s *samlTokenClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *samlTokenClient) Create(o *SamlToken) (*SamlToken, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*SamlToken), err
}

func (s *samlTokenClient) Get(name string, opts metav1.GetOptions) (*SamlToken, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*SamlToken), err
}

func (s *samlTokenClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*SamlToken, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*SamlToken), err
}

func (s *samlTokenClient) Update(o *SamlToken) (*SamlToken, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*SamlToken), err
}

func (s *samlTokenClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *samlTokenClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *samlTokenClient) List(opts metav1.ListOptions) (*SamlTokenList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*SamlTokenList), err
}

func (s *samlTokenClient) ListNamespaced(namespace string, opts metav1.ListOptions) (*SamlTokenList, error) {
	obj, err := s.objectClient.ListNamespaced(namespace, opts)
	return obj.(*SamlTokenList), err
}

func (s *samlTokenClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *samlTokenClient) Patch(o *SamlToken, patchType types.PatchType, data []byte, subresources ...string) (*SamlToken, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*SamlToken), err
}

func (s *samlTokenClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *samlTokenClient) AddHandler(ctx context.Context, name string, sync SamlTokenHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *samlTokenClient) AddFeatureHandler(ctx context.Context, enabled func() bool, name string, sync SamlTokenHandlerFunc) {
	s.Controller().AddFeatureHandler(ctx, enabled, name, sync)
}

func (s *samlTokenClient) AddLifecycle(ctx context.Context, name string, lifecycle SamlTokenLifecycle) {
	sync := NewSamlTokenLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *samlTokenClient) AddFeatureLifecycle(ctx context.Context, enabled func() bool, name string, lifecycle SamlTokenLifecycle) {
	sync := NewSamlTokenLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddFeatureHandler(ctx, enabled, name, sync)
}

func (s *samlTokenClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync SamlTokenHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *samlTokenClient) AddClusterScopedFeatureHandler(ctx context.Context, enabled func() bool, name, clusterName string, sync SamlTokenHandlerFunc) {
	s.Controller().AddClusterScopedFeatureHandler(ctx, enabled, name, clusterName, sync)
}

func (s *samlTokenClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle SamlTokenLifecycle) {
	sync := NewSamlTokenLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *samlTokenClient) AddClusterScopedFeatureLifecycle(ctx context.Context, enabled func() bool, name, clusterName string, lifecycle SamlTokenLifecycle) {
	sync := NewSamlTokenLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedFeatureHandler(ctx, enabled, name, clusterName, sync)
}
