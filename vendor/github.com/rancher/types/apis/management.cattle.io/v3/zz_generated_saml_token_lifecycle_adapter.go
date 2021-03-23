package v3

import (
	"github.com/rancher/norman/lifecycle"
	"github.com/rancher/norman/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

type SamlTokenLifecycle interface {
	Create(obj *SamlToken) (runtime.Object, error)
	Remove(obj *SamlToken) (runtime.Object, error)
	Updated(obj *SamlToken) (runtime.Object, error)
}

type samlTokenLifecycleAdapter struct {
	lifecycle SamlTokenLifecycle
}

func (w *samlTokenLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *samlTokenLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *samlTokenLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*SamlToken))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *samlTokenLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*SamlToken))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *samlTokenLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*SamlToken))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewSamlTokenLifecycleAdapter(name string, clusterScoped bool, client SamlTokenInterface, l SamlTokenLifecycle) SamlTokenHandlerFunc {
	if clusterScoped {
		resource.PutClusterScoped(SamlTokenGroupVersionResource)
	}
	adapter := &samlTokenLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *SamlToken) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
