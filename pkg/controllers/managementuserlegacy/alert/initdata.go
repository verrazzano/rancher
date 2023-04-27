// Copyright (c) 2023, Oracle and/or its affiliates.

// This file from the Rancher repository has been modified by Oracle as follows:
// - references to the cluster alerting CRDs and APIs have been removed
// - references to the cluster scanning CRDs and APIs have been removed

package alert

import (
	"fmt"

	v32 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/rancher/rancher/pkg/controllers/managementuserlegacy/alert/common"
	"github.com/rancher/rancher/pkg/controllers/managementuserlegacy/alert/manager"
	v3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	nodeAlertGroupName = "node-alert"

	// cluster scan related
	clusterScanAlertGroupName                    = "cluster-scan-alert"
	clusterScanManualAllAlertRuleName            = "cluster-scan-manual-all"
	clusterScanManualFailureOnlyAlertRuleName    = "cluster-scan-manual-failure-only"
	clusterScanScheduledAllAlertRuleName         = "cluster-scan-scheduled-all"
	clusterScanScheduledFailureOnlyAlertRuleName = "cluster-scan-scheduled--failure-only"
)

var (
	windowNodeLabel = labels.Set(map[string]string{"kubernetes.io/os": "windows"}).AsSelector()
	inherited       = false
)

type ProjectLifecycle struct {
	projectAlertGroups v3.ProjectAlertGroupInterface
	projectAlertRules  v3.ProjectAlertRuleInterface
	clusterName        string
}

// Create built-in project alerts
func (l *ProjectLifecycle) Create(obj *v3.Project) (runtime.Object, error) {
	name := "projectalert-workload-alert"
	projectName := obj.Name
	projectID := fmt.Sprintf("%s:%s", obj.Namespace, obj.Name)
	group := &v3.ProjectAlertGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: projectName,
		},
		Spec: v32.ProjectGroupSpec{
			ProjectName: projectID,
			CommonGroupField: v32.CommonGroupField{
				DisplayName: "A set of alerts for workload, pod, container",
				Description: "Alert for cpu, memory, disk, network",
				TimingField: defaultTimingField,
			},
		},
		Status: v32.AlertStatus{
			AlertState: "active",
		},
	}

	if _, err := l.projectAlertGroups.Create(group); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create built-in rules for deployment: %v", err)
	}

	name = "less-than-half-workload-available"
	rule := &v3.ProjectAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: projectName,
		},
		Spec: v32.ProjectAlertRuleSpec{
			ProjectName: projectID,
			GroupName:   common.GetGroupID(projectName, group.Name),
			CommonRuleField: v32.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: "Less than half workload available",
				TimingField: defaultTimingField,
			},
			WorkloadRule: &v32.WorkloadRule{
				Selector: map[string]string{
					"app": "workload",
				},
				AvailablePercentage: 50,
			},
		},
		Status: v32.AlertStatus{
			AlertState: "active",
		},
	}

	if _, err := l.projectAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "memory-close-to-resource-limited"
	rule = &v3.ProjectAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: projectName,
		},
		Spec: v32.ProjectAlertRuleSpec{
			ProjectName: projectID,
			GroupName:   common.GetGroupID(projectName, group.Name),
			CommonRuleField: v32.CommonRuleField{
				Severity:    SeverityWarning,
				DisplayName: "Memory usage close to the quota",
				TimingField: defaultTimingField,
			},
			MetricRule: &v32.MetricRule{
				Description:    "Container using memory close to the quota",
				Expression:     `sum(container_memory_working_set_bytes) by (pod_name, container_name) / sum(label_join(label_join(kube_pod_container_resource_limits_memory_bytes,"pod_name", "", "pod"),"container_name", "", "container")) by (pod_name, container_name)`,
				Comparison:     manager.ComparisonGreaterThan,
				Duration:       "3m",
				ThresholdValue: 1,
			},
		},
		Status: v32.AlertStatus{
			AlertState: "active",
		},
	}

	if _, err := l.projectAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}
	return obj, nil
}

func (l *ProjectLifecycle) Updated(obj *v3.Project) (runtime.Object, error) {
	return obj, nil
}

func (l *ProjectLifecycle) Remove(obj *v3.Project) (runtime.Object, error) {
	return obj, nil
}
