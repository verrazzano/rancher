// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package ocne

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	v1 "github.com/rancher/rancher/pkg/generated/norman/core/v1"
	"github.com/rancher/rancher/pkg/types/config"
	"net/http"
)

const (
	cmNamespace = "verrazzano-capi"
	cmName      = "ocne-metadata"
)

type handler struct {
	configmapLister v1.ConfigMapLister
}

func NewHandler(scaledContext *config.ScaledContext) http.Handler {
	return &handler{
		configmapLister: scaledContext.Core.ConfigMaps(cmNamespace).Controller().Lister(),
	}
}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	if request.Method != http.MethodGet {
		writeError(writer, http.StatusMethodNotAllowed, errors.New("Method not allowed"))
		return
	}

	ocneVersion := request.URL.Query().Get("ocneVersion")

	resource := mux.Vars(request)["resource"]
	switch resource {
	case "ocneVersions":
		handleRoute(writer, h.ocneVersions)
	case "metadata":
		handleRoute(writer, func() ([]byte, int, error) {
			return h.metadata(ocneVersion)
		})
	default:
		writeError(writer, http.StatusNotFound, errors.New("Not Found"))
	}
}

func handleRoute(writer http.ResponseWriter, f func() ([]byte, int, error)) {
	response, code, err := f()
	if err != nil {
		writeError(writer, code, err)
	}
	writer.Write(response)
}

func writeError(writer http.ResponseWriter, code int, err error) {
	writer.WriteHeader(code)
	message := fmt.Sprintf(`{"error": "%s"}`, err.Error())
	writer.Write([]byte(message))
}
