// Copyright (c) 2023, Oracle and/or its affiliates.

// This file from the Rancher repository has been modified by Oracle as follows:
// - references to the cluster cataloging CRDs and APIs have been removed

package manager

import (
	"testing"

	v32 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"

	"github.com/rancher/norman/types"
	v3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Set_Catalog_Error_State(t *testing.T) {
	type testcase struct {
		caseName       string
		catalogInfo    CatalogInfo
		catalog        v3.Catalog
		projectCatalog v3.ProjectCatalog
	}
	testcases := []testcase{
		{
			caseName: "default",
			catalogInfo: CatalogInfo{
				catalog:        nil,
				projectCatalog: nil,
			},
			catalog: v3.Catalog{
				ObjectMeta: v1.ObjectMeta{
					Name: "testCatalog"},
				Status: v32.CatalogStatus{
					LastRefreshTimestamp: "",
					Commit:               "",
					HelmVersionCommits:   nil,
					Conditions: []v32.CatalogCondition{
						{
							Type:    "Refreshed",
							Status:  "True",
							Message: "Test Catalog Stuff",
						},
					},
				},
			},
			projectCatalog: v3.ProjectCatalog{
				Namespaced:  types.Namespaced{},
				Catalog:     v3.Catalog{},
				ProjectName: "",
			},
		},
		{
			caseName: "catalogcondition nil status & nil message",
			catalogInfo: CatalogInfo{
				catalog:        nil,
				projectCatalog: nil,
			},
			catalog: v3.Catalog{
				ObjectMeta: v1.ObjectMeta{
					Name: "testCatalog"},
				Status: v32.CatalogStatus{
					LastRefreshTimestamp: "",
					Commit:               "",
					HelmVersionCommits:   nil,
				},
			},
			projectCatalog: v3.ProjectCatalog{
				Namespaced:  types.Namespaced{},
				Catalog:     v3.Catalog{},
				ProjectName: "",
			},
		},
		{
			caseName: "default",
			catalogInfo: CatalogInfo{
				catalog:        nil,
				projectCatalog: nil,
			},
			catalog: v3.Catalog{
				ObjectMeta: v1.ObjectMeta{
					Name: "testCatalog"},
				Status: v32.CatalogStatus{
					LastRefreshTimestamp: "",
					Commit:               "",
					HelmVersionCommits:   nil,
					Conditions: []v32.CatalogCondition{
						{
							Type: "Refreshed",
						},
					},
				},
			},
			projectCatalog: v3.ProjectCatalog{
				Namespaced:  types.Namespaced{},
				Catalog:     v3.Catalog{},
				ProjectName: "",
			},
		},
		{
			caseName: "false status",
			catalogInfo: CatalogInfo{
				catalog:        nil,
				projectCatalog: nil,
			},
			catalog: v3.Catalog{
				ObjectMeta: v1.ObjectMeta{
					Name: "testCatalog"},
				Status: v32.CatalogStatus{
					LastRefreshTimestamp: "",
					Commit:               "",
					HelmVersionCommits:   nil,
					Conditions: []v32.CatalogCondition{
						{
							Type:    "Refreshed",
							Status:  "False",
							Message: "Test Catalog Stuff",
						},
					},
				},
			},
			projectCatalog: v3.ProjectCatalog{
				Namespaced:  types.Namespaced{},
				Catalog:     v3.Catalog{},
				ProjectName: "",
			},
		},
		{
			caseName: "invalid status",
			catalogInfo: CatalogInfo{
				catalog:        nil,
				projectCatalog: nil,
			},
			catalog: v3.Catalog{
				ObjectMeta: v1.ObjectMeta{
					Name: "testCatalog"},
				Status: v32.CatalogStatus{
					LastRefreshTimestamp: "",
					Commit:               "",
					HelmVersionCommits:   nil,
					Conditions: []v32.CatalogCondition{
						{
							Type:    "Refreshed",
							Status:  "thisisnotanormalstatus",
							Message: "Test Catalog Stuff",
						},
					},
				},
			},
			projectCatalog: v3.ProjectCatalog{
				Namespaced:  types.Namespaced{},
				Catalog:     v3.Catalog{},
				ProjectName: "",
			},
		},
	}

	for _, c := range testcases {
		setCatalogErrorState(&c.catalogInfo, &c.catalog, &c.projectCatalog)
		assert.True(t, v32.CatalogConditionRefreshed.IsFalse(&c.catalog))
		assert.Equal(t, "Error syncing catalog testCatalog", v32.CatalogConditionRefreshed.GetMessage(&c.catalog))
	}
}
