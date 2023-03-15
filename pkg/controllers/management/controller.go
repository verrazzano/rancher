package management

import (
	"context"

	"github.com/verrazzano/rancher/pkg/clustermanager"
	"github.com/verrazzano/rancher/pkg/controllers/management/agentupgrade"
	"github.com/verrazzano/rancher/pkg/controllers/management/auth"
	"github.com/verrazzano/rancher/pkg/controllers/management/certsexpiration"
	"github.com/verrazzano/rancher/pkg/controllers/management/cloudcredential"
	"github.com/verrazzano/rancher/pkg/controllers/management/cluster"
	"github.com/verrazzano/rancher/pkg/controllers/management/clusterdeploy"
	"github.com/verrazzano/rancher/pkg/controllers/management/clustergc"
	"github.com/verrazzano/rancher/pkg/controllers/management/clusterprovisioner"
	"github.com/verrazzano/rancher/pkg/controllers/management/clusterstats"
	"github.com/verrazzano/rancher/pkg/controllers/management/clusterstatus"
	"github.com/verrazzano/rancher/pkg/controllers/management/clustertemplate"
	"github.com/verrazzano/rancher/pkg/controllers/management/drivers/kontainerdriver"
	"github.com/verrazzano/rancher/pkg/controllers/management/drivers/nodedriver"
	"github.com/verrazzano/rancher/pkg/controllers/management/etcdbackup"
	"github.com/verrazzano/rancher/pkg/controllers/management/kontainerdrivermetadata"
	"github.com/verrazzano/rancher/pkg/controllers/management/node"
	"github.com/verrazzano/rancher/pkg/controllers/management/nodepool"
	"github.com/verrazzano/rancher/pkg/controllers/management/nodetemplate"
	"github.com/verrazzano/rancher/pkg/controllers/management/podsecuritypolicy"
	"github.com/verrazzano/rancher/pkg/controllers/management/rbac"
	"github.com/verrazzano/rancher/pkg/controllers/management/restrictedadminrbac"
	"github.com/verrazzano/rancher/pkg/controllers/management/rkeworkerupgrader"
	"github.com/verrazzano/rancher/pkg/controllers/management/secretmigrator"
	"github.com/verrazzano/rancher/pkg/controllers/management/settings"
	"github.com/verrazzano/rancher/pkg/controllers/management/usercontrollers"
	"github.com/verrazzano/rancher/pkg/controllers/managementlegacy"
	"github.com/verrazzano/rancher/pkg/types/config"
	"github.com/verrazzano/rancher/pkg/wrangler"
)

func Register(ctx context.Context, management *config.ManagementContext, manager *clustermanager.Manager, wrangler *wrangler.Context) {
	// auth handlers need to run early to create namespaces that back clusters and projects
	// also, these handlers are purely in the mgmt plane, so they are lightweight compared to those that interact with machines and clusters
	auth.RegisterEarly(ctx, management, manager)
	usercontrollers.RegisterEarly(ctx, management, manager)

	// a-z
	agentupgrade.Register(ctx, management)
	certsexpiration.Register(ctx, management)
	cluster.Register(ctx, management)
	clusterdeploy.Register(ctx, management, manager)
	clustergc.Register(ctx, management)
	clusterprovisioner.Register(ctx, management)
	clusterstats.Register(ctx, management, manager)
	clusterstatus.Register(ctx, management)
	kontainerdriver.Register(ctx, management)
	kontainerdrivermetadata.Register(ctx, management)
	nodedriver.Register(ctx, management)
	nodepool.Register(ctx, management)
	cloudcredential.Register(ctx, management)
	node.Register(ctx, management, manager)
	podsecuritypolicy.Register(ctx, management)
	etcdbackup.Register(ctx, management)
	clustertemplate.Register(ctx, management)
	nodetemplate.Register(ctx, management)
	rkeworkerupgrader.Register(ctx, management, manager.ScaledContext)
	rbac.Register(ctx, management)
	restrictedadminrbac.Register(ctx, management, wrangler)
	secretmigrator.Register(ctx, management)
	settings.Register(ctx, management)
	managementlegacy.Register(ctx, management, manager)

	// Ensure caches are available for user controllers, these are used as part of
	// registration
	management.Management.ClusterAlertGroups("").Controller()
	management.Management.ClusterAlertRules("").Controller()

	// Register last
	auth.RegisterLate(ctx, management)
}
