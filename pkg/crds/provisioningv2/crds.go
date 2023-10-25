package provisioningv2

import (
	"embed"

	v1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/rancher/pkg/features"
	"github.com/rancher/wrangler/pkg/crd"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/rancher/wrangler/pkg/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	capiCRDs = map[string]bool{
		"Machine":            true,
		"MachineHealthCheck": true,
		"MachineDeployment":  true,
		"MachineSet":         true,
		"Cluster":            true,
	}

	//go:embed capi-crds.yaml capi-webhooks.yaml
	capiData embed.FS
)

func List() (result []crd.CRD) {
	result = append(result, provisioning()...)
	if features.EmbeddedClusterAPI.Enabled() {
		result = append(result, capi()...)
	}
	return
}

func provisioning() []crd.CRD {
	return []crd.CRD{
		newRancherCRD(&v1.Cluster{}, func(c crd.CRD) crd.CRD {
			c.Labels = map[string]string{
				"auth.cattle.io/cluster-indexed": "true",
			}
			return c.
				WithColumn("Ready", ".status.ready").
				WithColumn("Kubeconfig", ".status.clientSecretName")
		}),
	}
}

func Webhooks() []runtime.Object {
	if features.EmbeddedClusterAPI.Enabled() {
		return capiWebhooks()
	}
	return nil
}

func capiWebhooks() []runtime.Object {
	f, err := capiData.Open("capi-webhooks.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	objs, err := yaml.ToObjects(f)
	if err != nil {
		panic(err)
	}

	return objs
}

func capi() []crd.CRD {
	f, err := capiData.Open("capi-crds.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	objs, err := yaml.ToObjects(f)
	if err != nil {
		panic(err)
	}

	var result []crd.CRD
	for _, obj := range objs {
		if obj.GetObjectKind().GroupVersionKind().Kind != "CustomResourceDefinition" {
			continue
		}
		if unstr, ok := obj.(*unstructured.Unstructured); ok &&
			capiCRDs[data.Object(unstr.Object).String("spec", "names", "kind")] {
			labels := unstr.GetLabels()
			if labels == nil {
				labels = map[string]string{}
			}
			labels["auth.cattle.io/cluster-indexed"] = "true"
			unstr.SetLabels(labels)
			result = append(result, crd.CRD{
				Override: obj,
			})
		}
	}

	return result
}

func newRancherCRD(obj interface{}, customize func(crd.CRD) crd.CRD) crd.CRD {
	crd := crd.CRD{
		GVK: schema.GroupVersionKind{
			Group:   "provisioning.cattle.io",
			Version: "v1",
		},
		Status:       true,
		SchemaObject: obj,
	}
	if customize != nil {
		crd = customize(crd)
	}
	return crd
}
