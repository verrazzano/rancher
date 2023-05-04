// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package ocne

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"net/http"
)

func (h *handler) kubernetesVersions() ([]byte, int, error) {
	cm, err := h.configmapLister.Get(cmNamespace, cmName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return []byte("{}"), http.StatusOK, nil
		}
		return nil, http.StatusInternalServerError, err
	}

	data, err := apiyaml.ToJSON([]byte(cm.Data["mapping"]))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return data, http.StatusOK, nil
}
