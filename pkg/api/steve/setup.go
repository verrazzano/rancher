package steve

import (
	"context"
	steve "github.com/rancher/steve/pkg/server"
	"github.com/verrazzano/rancher/pkg/api/steve/catalog"
	"github.com/verrazzano/rancher/pkg/api/steve/clusters"
	"github.com/verrazzano/rancher/pkg/api/steve/disallow"
	"github.com/verrazzano/rancher/pkg/api/steve/navlinks"
	"github.com/verrazzano/rancher/pkg/api/steve/settings"
	"github.com/verrazzano/rancher/pkg/api/steve/userpreferences"
	"github.com/verrazzano/rancher/pkg/wrangler"
)

func Setup(ctx context.Context, server *steve.Server, config *wrangler.Context) error {
	userpreferences.Register(server.BaseSchemas, server.ClientFactory)
	if err := clusters.Register(ctx, server, config); err != nil {
		return err
	}
	navlinks.Register(ctx, server)
	settings.Register(server)
	disallow.Register(server)
	return catalog.Register(ctx,
		server,
		config.HelmOperations,
		config.CatalogContentManager)
}
