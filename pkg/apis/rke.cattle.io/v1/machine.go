package v1

import (
	"github.com/rancher/wrangler/pkg/genericcondition"
	corev1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

type RKECommonNodeConfig struct {
	Labels                    map[string]string `json:"labels,omitempty"`
	Taints                    []corev1.Taint    `json:"taints,omitempty"`
	CloudCredentialSecretName string            `json:"cloudCredentialSecretName,omitempty"`
}

type RKEMachineStatus struct {
	Conditions                []genericcondition.GenericCondition `json:"conditions,omitempty"`
	JobComplete               bool                                `json:"jobComplete,omitempty"`
	JobName                   string                              `json:"jobName,omitempty"`
	Ready                     bool                                `json:"ready,omitempty"`
	DriverHash                string                              `json:"driverHash,omitempty"`
	DriverURL                 string                              `json:"driverUrl,omitempty"`
	CloudCredentialSecretName string                              `json:"cloudCredentialSecretName,omitempty"`
	FailureReason             string                              `json:"failureReason,omitempty"`
	FailureMessage            string                              `json:"failureMessage,omitempty"`
	Addresses                 []capi.MachineAddress               `json:"addresses,omitempty"`
}
