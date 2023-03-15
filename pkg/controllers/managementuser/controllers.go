package managementuser

import (
	"context"

	"github.com/verrazzano/rancher/pkg/controllers/managementlegacy/compose/common"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/certsexpiration"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/clusterauthtoken"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/healthsyncer"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/machinerole"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/networkpolicy"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/nodesyncer"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/nsserviceaccount"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/pspdelete"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/rbac"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/rbac/podsecuritypolicy"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/resourcequota"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/secret"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/snapshotbackpopulate"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/windows"
	"github.com/verrazzano/rancher/pkg/controllers/managementuserlegacy"
	"github.com/verrazzano/rancher/pkg/features"
	managementv3 "github.com/verrazzano/rancher/pkg/generated/norman/management.cattle.io/v3"
	"github.com/verrazzano/rancher/pkg/impersonation"
	"github.com/verrazzano/rancher/pkg/types/config"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Register(ctx context.Context, mgmt *config.ScaledContext, cluster *config.UserContext, clusterRec *managementv3.Cluster, kubeConfigGetter common.KubeConfigGetter) error {
	rbac.Register(ctx, cluster)
	healthsyncer.Register(ctx, cluster)
	networkpolicy.Register(ctx, cluster)
	nodesyncer.Register(ctx, cluster, kubeConfigGetter)
	podsecuritypolicy.Register(ctx, cluster)
	secret.Register(ctx, cluster)
	resourcequota.Register(ctx, cluster)
	certsexpiration.Register(ctx, cluster)
	windows.Register(ctx, clusterRec, cluster)
	nsserviceaccount.Register(ctx, cluster)
	if features.RKE2.Enabled() {
		snapshotbackpopulate.Register(ctx, cluster)
		pspdelete.Register(ctx, cluster)
		machinerole.Register(ctx, cluster)
	}

	// register controller for API
	cluster.APIAggregation.APIServices("").Controller()
	// register secrets controller for impersonation
	cluster.Core.Secrets("").Controller()

	if clusterRec.Spec.LocalClusterAuthEndpoint.Enabled {
		err := clusterauthtoken.CRDSetup(ctx, cluster.UserOnlyContext())
		if err != nil {
			return err
		}
		clusterauthtoken.Register(ctx, cluster)
	}

	// Ensure these caches are started
	cluster.Core.Namespaces("").Controller()
	cluster.Core.Secrets("").Controller()
	cluster.Core.ServiceAccounts("").Controller()

	return managementuserlegacy.Register(ctx, mgmt, cluster, clusterRec, kubeConfigGetter)
}

func RegisterFollower(ctx context.Context, cluster *config.UserContext, kubeConfigGetter common.KubeConfigGetter, clusterManager healthsyncer.ClusterControllerLifecycle) error {
	cluster.KindNamespaces[schema.GroupVersionKind{
		Version: "v1",
		Kind:    "Secret",
	}] = impersonation.ImpersonationNamespace
	cluster.KindNamespaces[schema.GroupVersionKind{
		Version: "v1",
		Kind:    "ServiceAccount",
	}] = impersonation.ImpersonationNamespace

	cluster.Core.Namespaces("").Controller()
	cluster.Core.Secrets("").Controller()
	cluster.Core.ServiceAccounts("").Controller()
	cluster.RBAC.ClusterRoleBindings("").Controller()
	cluster.RBAC.ClusterRoles("").Controller()
	cluster.RBAC.RoleBindings("").Controller()
	cluster.RBAC.Roles("").Controller()
	return nil
}
