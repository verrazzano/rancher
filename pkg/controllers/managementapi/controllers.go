package managementapi

import (
	"context"

	normanapi "github.com/rancher/norman/api"
	"github.com/verrazzano/rancher/pkg/auth/tokens"
	"github.com/verrazzano/rancher/pkg/clustermanager"
	"github.com/verrazzano/rancher/pkg/controllers/management/auth"
	v3cluster "github.com/verrazzano/rancher/pkg/controllers/management/cluster"
	podsecuritypolicy2 "github.com/verrazzano/rancher/pkg/controllers/management/podsecuritypolicy"
	"github.com/verrazzano/rancher/pkg/controllers/managementapi/catalog"
	"github.com/verrazzano/rancher/pkg/controllers/managementapi/dynamicschema"
	"github.com/verrazzano/rancher/pkg/controllers/managementapi/samlconfig"
	"github.com/verrazzano/rancher/pkg/controllers/managementapi/usercontrollers"
	whitelistproxyKontainerDriver "github.com/verrazzano/rancher/pkg/controllers/managementapi/whitelistproxy/kontainerdriver"
	whitelistproxyNodeDriver "github.com/verrazzano/rancher/pkg/controllers/managementapi/whitelistproxy/nodedriver"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/clusterauthtoken"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/rbac"
	"github.com/verrazzano/rancher/pkg/controllers/managementuser/rbac/podsecuritypolicy"
	"github.com/verrazzano/rancher/pkg/controllers/managementuserlegacy/monitoring"
	"github.com/verrazzano/rancher/pkg/types/config"
)

func Register(ctx context.Context, scaledContext *config.ScaledContext, clusterManager *clustermanager.Manager, server *normanapi.Server) error {
	if err := registerIndexers(scaledContext); err != nil {
		return err
	}

	catalog.Register(ctx, scaledContext)
	dynamicschema.Register(ctx, scaledContext, server.Schemas)
	whitelistproxyNodeDriver.Register(ctx, scaledContext)
	whitelistproxyKontainerDriver.Register(ctx, scaledContext)
	samlconfig.Register(ctx, scaledContext)
	usercontrollers.Register(ctx, scaledContext, clusterManager)
	return nil
}

func registerIndexers(scaledContext *config.ScaledContext) error {
	if err := clusterauthtoken.RegisterIndexers(scaledContext); err != nil {
		return err
	}
	if err := rbac.RegisterIndexers(scaledContext); err != nil {
		return err
	}
	if err := monitoring.RegisterIndexers(scaledContext); err != nil {
		return err
	}
	if err := auth.RegisterIndexers(scaledContext); err != nil {
		return err
	}
	if err := tokens.RegisterIndexer(scaledContext); err != nil {
		return err
	}
	if err := podsecuritypolicy.RegisterIndexers(scaledContext); err != nil {
		return err
	}
	if err := podsecuritypolicy2.RegisterIndexers(scaledContext); err != nil {
		return err
	}
	v3cluster.RegisterIndexers(scaledContext)
	return nil
}
