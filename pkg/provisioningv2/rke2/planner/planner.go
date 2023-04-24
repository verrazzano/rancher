package planner

import (
	"context"
	"fmt"
	"github.com/moby/locker"
	"github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1/plan"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2"
	capicontrollers "github.com/rancher/rancher/pkg/generated/controllers/cluster.x-k8s.io/v1beta1"
	mgmtcontrollers "github.com/rancher/rancher/pkg/generated/controllers/management.cattle.io/v3"
	ranchercontrollers "github.com/rancher/rancher/pkg/generated/controllers/provisioning.cattle.io/v1"
	rkecontrollers "github.com/rancher/rancher/pkg/generated/controllers/rke.cattle.io/v1"
	corecontrollers "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sort"
	"strings"
)

const (
	ConfigYamlFileName = "/etc/rancher/%s/config.yaml.d/50-rancher.yaml"
)

type ErrWaiting string

func (e ErrWaiting) Error() string {
	return string(e)
}

type Planner struct {
	ctx                           context.Context
	store                         *PlanStore
	etcdSnapshotCache             rkecontrollers.ETCDSnapshotCache
	secretClient                  corecontrollers.SecretClient
	secretCache                   corecontrollers.SecretCache
	machines                      capicontrollers.MachineClient
	machinesCache                 capicontrollers.MachineCache
	clusterRegistrationTokenCache mgmtcontrollers.ClusterRegistrationTokenCache
	capiClient                    capicontrollers.ClusterClient
	capiClusters                  capicontrollers.ClusterCache
	managementClusters            mgmtcontrollers.ClusterCache
	rancherClusterCache           ranchercontrollers.ClusterCache
	locker                        locker.Locker
}

func (p *Planner) setMachineConditionStatus(clusterPlan *plan.Plan, machineNames []string, messagePrefix string, messages map[string][]string) error {
	var waiting bool
	for _, machineName := range machineNames {
		machine := clusterPlan.Machines[machineName]
		if machine == nil {
			return fmt.Errorf("found unexpected machine %s that is not in cluster plan", machineName)
		}

		if !rke2.InfrastructureReady.IsTrue(machine) {
			waiting = true
			continue
		}

		machine = machine.DeepCopy()
		if message := messages[machineName]; len(message) > 0 {
			msg := strings.Join(message, ", ")
			waiting = true
			if rke2.Reconciled.GetMessage(machine) == msg {
				continue
			}
			conditions.MarkUnknown(machine, capi.ConditionType(rke2.Reconciled), "Waiting", msg)
		} else if !rke2.Reconciled.IsTrue(machine) {
			// Since there is no status message, then the condition should be set to true.
			conditions.MarkTrue(machine, capi.ConditionType(rke2.Reconciled))

			// Even though we are technically not waiting for something, an error should be returned so that the planner will retry.
			// The machine being updated will cause the planner to re-enqueue with the new data.
			waiting = true
		} else {
			continue
		}

		if _, err := p.machines.UpdateStatus(machine); err != nil {
			return err
		}
	}

	if waiting {
		return ErrWaiting(messagePrefix + atMostThree(machineNames) + detailMessage(machineNames, messages))
	}
	return nil
}

func atMostThree(names []string) string {
	sort.Strings(names)
	if len(names) > 3 {
		return fmt.Sprintf("%s and %d more", strings.Join(names[:3], ","), len(names)-3)
	}
	return strings.Join(names, ",")
}

func detailMessage(machines []string, messages map[string][]string) string {
	if len(machines) != 1 {
		return ""
	}
	message := messages[machines[0]]
	if len(message) != 0 {
		return fmt.Sprintf(": %s", strings.Join(message, ", "))
	}
	return ""
}

func isEtcd(entry *planEntry) bool {
	return entry.Metadata != nil && entry.Metadata.Labels[rke2.EtcdRoleLabel] == "true"
}

func isControlPlane(entry *planEntry) bool {
	return entry.Metadata != nil && entry.Metadata.Labels[rke2.ControlPlaneRoleLabel] == "true"
}

func isWorker(entry *planEntry) bool {
	return entry.Metadata != nil && entry.Metadata.Labels[rke2.WorkerRoleLabel] == "true"
}

type planEntry struct {
	Machine  *capi.Machine
	Plan     *plan.Node
	Metadata *plan.Metadata
}
