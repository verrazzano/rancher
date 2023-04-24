// Copyright (c) 2023, Oracle and/or its affiliates.

// This file from the Rancher repository has been modified by Oracle as follows:
// - references to the cluster catalog CRDs and APIs have been removed

package manager

import (
	v3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
)

type CatalogInfo struct {
	catalog        *v3.Catalog
	projectCatalog *v3.ProjectCatalog
}
