// Copyright (c) 2023, Oracle and/or its affiliates.

// This file from the Rancher repository has been modified by Oracle as follows:
// - references to the cluster alerting CRDs and APIs have been removed

package alert

import (
	"fmt"

	v32 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	v3client "github.com/rancher/rancher/pkg/client/generated/management/v3"
)

const monitoringEnabled = "MonitoringEnabled"

func ProjectAlertRuleValidator(resquest *types.APIContext, schema *types.Schema, data map[string]interface{}) error {
	projectID := data["projectId"].(string)

	var spec v32.ProjectAlertRuleSpec
	if err := convert.ToObj(data, &spec); err != nil {
		return httperror.NewAPIError(httperror.InvalidBodyContent, fmt.Sprintf("%v", err))
	}

	if spec.MetricRule != nil {
		project := &v3client.Project{}
		if err := access.ByID(resquest, resquest.Version, v3client.ProjectType, projectID, project); err != nil {
			return fmt.Errorf("access project by id failed, %v", err)
		}
		if project.Conditions != nil {
			for _, v := range project.Conditions {
				if v.Type == monitoringEnabled && v.Status == "True" {
					return nil
				}
			}
		}
		return fmt.Errorf("if you want to use metric alert, need to enable monitoring for project %s", projectID)
	}

	return nil
}
