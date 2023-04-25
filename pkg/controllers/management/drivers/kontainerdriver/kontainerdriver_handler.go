// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package kontainerdriver

import (
	"net/http"
)

type Handler struct {
	fileHandler http.Handler
}

func NewKontainerDriverHandler() http.Handler {
	return &Handler{
		fileHandler: http.FileServer(http.Dir("/var/lib/rancher-data/drivers")),
	}
}

func (k *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	k.fileHandler.ServeHTTP(writer, req)
}
