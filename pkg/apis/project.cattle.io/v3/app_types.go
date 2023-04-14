package v3

import (
	"github.com/rancher/norman/condition"
	v1 "k8s.io/api/core/v1"
)

var (
	AppConditionInstalled           condition.Cond = "Installed"
	AppConditionMigrated            condition.Cond = "Migrated"
	AppConditionDeployed            condition.Cond = "Deployed"
	AppConditionForceUpgrade        condition.Cond = "ForceUpgrade"
	AppConditionUserTriggeredAction condition.Cond = "UserTriggeredAction"
)

type AppCondition struct {
	// Type of cluster condition.
	Type condition.Cond `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition
	Message string `json:"message,omitempty"`
}
