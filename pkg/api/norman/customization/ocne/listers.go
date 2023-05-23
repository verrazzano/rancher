// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package ocne

import (
	"encoding/json"
	"helm.sh/helm/v3/pkg/repo"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"net/http"
	"sort"
)

type Version struct {
	Release         string `json:"Release"`
	ContainerImages struct {
		Calico         string `json:"calico"`
		Coredns        string `json:"coredns"`
		Etcd           string `json:"etcd"`
		TigeraOperator string `json:"tigera-operator"`
	} `json:"container-images"`
}

func (h *handler) loadVersionMapping() (map[string]Version, error) {
	data, err := h.getOCNEMetadataJSON()
	if err != nil {
		return nil, err
	}

	versionMapping := map[string]Version{}
	if err := json.Unmarshal(data, &versionMapping); err != nil {
		return nil, err
	}
	return versionMapping, nil
}

func (h *handler) loadModulesMetadata() (map[string]repo.ChartVersions, error) {
	data, err := h.getModulesMetadataJSON()
	if err != nil {
		return nil, err
	}

	modulesMetadata := map[string]repo.ChartVersions{}
	if err := json.Unmarshal(data, &modulesMetadata); err != nil {
		return nil, err
	}
	return modulesMetadata, nil
}

func (h *handler) getOCNEMetadataJSON() ([]byte, error) {
	cm, err := h.configmapLister.Get(cmNamespace, ocneCmName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return []byte("{}"), nil
		}
		return nil, err
	}

	data, err := apiyaml.ToJSON([]byte(cm.Data["mapping"]))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (h *handler) getModulesMetadataJSON() ([]byte, error) {
	cm, err := h.configmapLister.Get(cmNamespace, modulesCmName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return []byte("{}"), nil
		}
		return nil, err
	}

	data, err := apiyaml.ToJSON([]byte(cm.Data["entries"]))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (h *handler) metadata(ocneVersion string) ([]byte, int, error) {
	versionMapping, err := h.loadVersionMapping()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	etcdTags := sortedVersionHelper(versionMapping, func(version Version) string {
		return version.ContainerImages.Etcd
	}, ocneVersion)
	coreDNSTags := sortedVersionHelper(versionMapping, func(version Version) string {
		return version.ContainerImages.Coredns
	}, ocneVersion)
	tigeraTags := sortedVersionHelper(versionMapping, func(version Version) string {
		return version.ContainerImages.TigeraOperator
	}, ocneVersion)

	var kubernetesVersions []string
	for k, v := range versionMapping {
		if v.Release == ocneVersion {
			kubernetesVersions = append(kubernetesVersions, k)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(kubernetesVersions)))

	result := map[string][]string{}
	result["etcd"] = etcdTags
	result["coredns"] = coreDNSTags
	result["tigeraOperator"] = tigeraTags
	result["kubernetesVersions"] = kubernetesVersions

	data, err := json.Marshal(result)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return data, http.StatusOK, nil
}

func (h *handler) modules() ([]byte, int, error) {
	modulesMetadata, err := h.loadModulesMetadata()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	data, err := json.Marshal(modulesMetadata)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return data, http.StatusOK, nil
}

func (h *handler) ocneVersions() ([]byte, int, error) {
	versionMapping, err := h.loadVersionMapping()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	data, err := h.buildSortedListFromVersionMapping(versionMapping, func(version Version) string {
		return version.Release
	}, func(strings []string) {
		sort.Sort(sort.Reverse(sort.StringSlice(strings)))
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return data, http.StatusOK, nil
}

func (h *handler) buildSortedListFromVersionMapping(versionMapping map[string]Version, accessor func(version Version) string, sorter func([]string)) ([]byte, error) {
	items := sortedItems(versionMapping, accessor, sorter)
	data, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func sortedVersionHelper(versionMapping map[string]Version, accessor func(version Version) string, ocneVersion string) []string {
	return sortedItems(versionMapping, func(version Version) string {
		if version.Release == ocneVersion {
			return accessor(version)
		}
		return ""
	}, func(strings []string) {
		sort.Sort(sort.Reverse(sort.StringSlice(strings)))
	})
}

func sortedItems(versionMapping map[string]Version, accessor func(version Version) string, sorter func([]string)) []string {
	items := map[string]bool{}
	for _, v := range versionMapping {
		val := accessor(v)
		if val != "" {
			items[accessor(v)] = true
		}
	}

	var list []string
	for k := range items {
		list = append(list, k)
	}
	sorter(list)

	return list
}
