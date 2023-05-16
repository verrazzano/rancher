package provisioningv2

import (
	"context"

	"github.com/rancher/rancher/pkg/controllers/provisioningv2/cluster"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/fleetcluster"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/fleetworkspace"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/managedchart"
	"github.com/rancher/rancher/pkg/features"
	"github.com/rancher/rancher/pkg/provisioningv2/kubeconfig"
	"github.com/rancher/rancher/pkg/wrangler"
)

func Register(ctx context.Context, clients *wrangler.Context) error {
	kubeconfigManager := kubeconfig.New(clients)
	cluster.Register(ctx, clients, kubeconfigManager)

	if features.Fleet.Enabled() {
		managedchart.Register(ctx, clients)
		fleetcluster.Register(ctx, clients)
		fleetworkspace.Register(ctx, clients)
	}

	return nil
}
