package networkpolicy

import (
	"github.com/sirupsen/logrus"
	v3 "github.com/verrazzano/rancher/pkg/generated/norman/management.cattle.io/v3"
	"k8s.io/apimachinery/pkg/runtime"
)

type projectNetworkPolicySyncer struct {
	npmgr *netpolMgr
}

// Sync invokes the Policy Handler to take care of installing the native network policies
func (pnps *projectNetworkPolicySyncer) Sync(key string, pnp *v3.ProjectNetworkPolicy) (runtime.Object, error) {
	if pnp == nil || pnp.DeletionTimestamp != nil {
		return nil, nil
	}
	logrus.Debugf("projectNetworkPolicySyncer: Sync: pnp=%+v", pnp)
	return nil, pnps.npmgr.programNetworkPolicy(pnp.Namespace, pnps.npmgr.clusterNamespace)
}
