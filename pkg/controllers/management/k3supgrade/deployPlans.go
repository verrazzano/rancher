package k3supgrade

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	planv1 "github.com/rancher/system-upgrade-controller/pkg/apis/upgrade.cattle.io/v1"
	planClientset "github.com/rancher/system-upgrade-controller/pkg/generated/clientset/versioned/typed/upgrade.cattle.io/v1"
	v3 "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const MaxDisplayNodes = 10

// deployPlans creates a master and worker plan in the downstream cluster to instrument
// the system-upgrade-controller in the downstream cluster
func (h *handler) deployPlans(cluster *v3.Cluster) error {

	// access downstream cluster
	clusterCtx, err := h.manager.UserContext(cluster.Name)
	if err != nil {
		return err

	}

	// create a client for GETing Plans in the downstream cluster
	planConfig, err := planClientset.NewForConfig(&clusterCtx.RESTConfig)
	if err != nil {
		return err
	}
	planClient := planConfig.Plans(metav1.NamespaceAll)

	planList, err := planClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	masterPlan := &planv1.Plan{}
	workerPlan := &planv1.Plan{}
	// deactivate all existing plans that are not managed by Rancher
	for i, plan := range planList.Items {
		if _, ok := plan.Labels[rancherManagedPlan]; !ok {
			// inverse selection is used here, we select a non-existent label
			lsr := metav1.LabelSelectorRequirement{
				Key:      upgradeDisableLabelKey,
				Operator: metav1.LabelSelectorOpExists,
			}
			plan.Spec.NodeSelector.MatchExpressions = append(plan.Spec.NodeSelector.MatchExpressions, lsr)

			_, err = planClient.Update(context.TODO(), &plan, metav1.UpdateOptions{})
			if err != nil {
				return err
			}

		} else {

			switch name := plan.Name; name {
			case k3sMasterPlanName:
				if plan.Namespace == systemUpgradeNS {
					// reference absolute memory location
					masterPlan = &planList.Items[i]
				}
			case k3sWorkerPlanName:
				if plan.Namespace == systemUpgradeNS {
					// reference absolute memory location
					workerPlan = &planList.Items[i]
				}
			}
		}
	}
	// if rancher plans exist, do we need to update?
	if masterPlan.Name != "" || workerPlan.Name != "" {
		if masterPlan.Name != "" {
			newMaster := configureMasterPlan(*masterPlan, cluster.Spec.K3sConfig.Version,
				cluster.Spec.K3sConfig.ServerConcurrency,
				cluster.Spec.K3sConfig.DrainServerNodes)

			if !cmp(*masterPlan, newMaster) {
				planClient = planConfig.Plans(systemUpgradeNS)
				masterPlan, err = planClient.Update(context.TODO(), &newMaster, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
			}
		}

		if workerPlan.Name != "" {
			newWorker := configureWorkerPlan(*workerPlan, cluster.Spec.K3sConfig.Version,
				cluster.Spec.K3sConfig.WorkerConcurrency,
				cluster.Spec.K3sConfig.DrainWorkerNodes)

			if !cmp(*workerPlan, newWorker) {
				planClient = planConfig.Plans(systemUpgradeNS)
				workerPlan, err = planClient.Update(context.TODO(), &newWorker, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
			}
		}

	} else { // create the plans
		planClient = planConfig.Plans(systemUpgradeNS)
		genMasterPlan := generateMasterPlan(cluster.Spec.K3sConfig.Version,
			cluster.Spec.K3sConfig.ServerConcurrency,
			cluster.Spec.K3sConfig.DrainServerNodes)

		masterPlan, err = planClient.Create(context.TODO(), &genMasterPlan, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		genWorkerPlan := generateWorkerPlan(cluster.Spec.K3sConfig.Version,
			cluster.Spec.K3sConfig.WorkerConcurrency,
			cluster.Spec.K3sConfig.DrainWorkerNodes)

		workerPlan, err = planClient.Create(context.TODO(), &genWorkerPlan, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		logrus.Infof("Plans successfully deployed into cluster %s", cluster.Name)
	}

	cluster, err = h.modifyClusterCondition(cluster, *masterPlan, *workerPlan)
	if err != nil {
		return err
	}
	return nil
}

//cmp compares two plans but does not compare their Status, returns true if they are the same
func cmp(a, b planv1.Plan) bool {
	if a.Namespace != b.Namespace {
		return false
	}

	if a.Spec.Version != b.Spec.Version {
		return false
	}

	if a.Spec.Concurrency != b.Spec.Concurrency {
		return false
	}

	if !reflect.DeepEqual(a.Spec, b.Spec) {
		return false
	}
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return false
	}
	if !reflect.DeepEqual(a.TypeMeta, b.TypeMeta) {
		return false
	}
	return true
}

//cluster state management during the upgrade, plans may be ""
func (h *handler) modifyClusterCondition(cluster *v3.Cluster, masterPlan, workerPlan planv1.Plan) (*v3.Cluster, error) {

	// implement a simple state machine
	// UpgradedTrue => GenericUpgrading =>  MasterPlanUpgrading || WorkerPlanUpgrading =>  UpgradedTrue

	if masterPlan.Name == "" && workerPlan.Name == "" {
		// enter upgrading state
		if v3.ClusterConditionUpgraded.IsTrue(cluster) {
			v3.ClusterConditionUpgraded.Unknown(cluster)
			v3.ClusterConditionUpgraded.Message(cluster, "cluster is being upgraded")
			return h.clusterClient.Update(cluster)
		}
		if v3.ClusterConditionUpgraded.IsUnknown(cluster) {
			// remain in upgrading state if we are passed empty plans
			return cluster, nil
		}
	}

	if masterPlan.Name != "" && len(masterPlan.Status.Applying) > 0 {
		v3.ClusterConditionUpgraded.Unknown(cluster)
		c := cluster.Spec.K3sConfig.ServerConcurrency
		masterPlanMessage := fmt.Sprintf("controlplane node [%s] being upgraded",
			upgradingMessage(c, masterPlan.Status.Applying))
		return h.enqueueOrUpdate(cluster, masterPlanMessage)

	}

	if workerPlan.Name != "" && len(workerPlan.Status.Applying) > 0 {
		v3.ClusterConditionUpgraded.Unknown(cluster)
		c := cluster.Spec.K3sConfig.WorkerConcurrency
		workerPlanMessage := fmt.Sprintf("worker node [%s] being upgraded",
			upgradingMessage(c, workerPlan.Status.Applying))
		return h.enqueueOrUpdate(cluster, workerPlanMessage)
	}

	// if we made it this far nothing is applying
	// see k3supgrade_handler also
	v3.ClusterConditionUpgraded.True(cluster)
	return h.clusterClient.Update(cluster)

}

func upgradingMessage(concurrency int, nodes []string) string {
	// concurrency max can be very large
	if concurrency > len(nodes) {
		concurrency = len(nodes)
	}
	if concurrency > MaxDisplayNodes {
		concurrency = MaxDisplayNodes
	}

	return strings.Join(nodes[:concurrency], ", ")
}

func (h *handler) enqueueOrUpdate(cluster *v3.Cluster, upgradeMessage string) (*v3.Cluster, error) {
	if v3.ClusterConditionUpgraded.GetMessage(cluster) == upgradeMessage {
		// update would be no op
		h.clusterEnqueueAfter(cluster.Name, time.Second*5) // prevent controller from remaining in this state
		return cluster, nil
	}

	v3.ClusterConditionUpgraded.Message(cluster, upgradeMessage)
	return h.clusterClient.Update(cluster)

}
