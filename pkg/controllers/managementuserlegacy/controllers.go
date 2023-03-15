package managementuserlegacy

import (
	"context"

	"github.com/verrazzano/rancher/pkg/controllers/managementlegacy/compose/common"
	"github.com/verrazzano/rancher/pkg/controllers/managementuserlegacy/alert"
	"github.com/verrazzano/rancher/pkg/controllers/managementuserlegacy/approuter"
	"github.com/verrazzano/rancher/pkg/controllers/managementuserlegacy/globaldns"
	"github.com/verrazzano/rancher/pkg/controllers/managementuserlegacy/helm"
	"github.com/verrazzano/rancher/pkg/controllers/managementuserlegacy/monitoring"
	"github.com/verrazzano/rancher/pkg/controllers/managementuserlegacy/systemimage"
	managementv3 "github.com/verrazzano/rancher/pkg/generated/norman/management.cattle.io/v3"
	"github.com/verrazzano/rancher/pkg/types/config"
)

func Register(ctx context.Context, mgmt *config.ScaledContext, cluster *config.UserContext, clusterRec *managementv3.Cluster, kubeConfigGetter common.KubeConfigGetter) error {
	helm.Register(ctx, mgmt, cluster, kubeConfigGetter)
	systemimage.Register(ctx, cluster)
	approuter.Register(ctx, cluster)
	alert.Register(ctx, mgmt, cluster)
	globaldns.Register(ctx, cluster)
	monitoring.Register(ctx, cluster)

	// register controller for API
	cluster.APIAggregation.APIServices("").Controller()
	return nil
}
