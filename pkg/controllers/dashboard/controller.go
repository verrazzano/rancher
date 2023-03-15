package dashboard

import (
	"context"

	"github.com/rancher/wrangler/pkg/needacert"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/apiservice"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/clusterindex"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/clusterregistrationtoken"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/cspadaptercharts"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/fleetcharts"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/helm"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/hostedcluster"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/kubernetesprovider"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/mcmagent"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/scaleavailable"
	"github.com/verrazzano/rancher/pkg/controllers/dashboard/systemcharts"
	"github.com/verrazzano/rancher/pkg/controllers/management/clusterconnected"
	"github.com/verrazzano/rancher/pkg/controllers/provisioningv2"
	"github.com/verrazzano/rancher/pkg/features"
	"github.com/verrazzano/rancher/pkg/wrangler"
)

func Register(ctx context.Context, wrangler *wrangler.Context, embedded bool) error {
	helm.Register(ctx, wrangler)
	kubernetesprovider.Register(ctx,
		wrangler.Mgmt.Cluster(),
		wrangler.K8s,
		wrangler.MultiClusterManager)
	apiservice.Register(ctx, wrangler, embedded)
	needacert.Register(ctx,
		wrangler.Core.Secret(),
		wrangler.Core.Service(),
		wrangler.Admission.MutatingWebhookConfiguration(),
		wrangler.Admission.ValidatingWebhookConfiguration(),
		wrangler.CRD.CustomResourceDefinition())
	scaleavailable.Register(ctx, wrangler)
	if err := systemcharts.Register(ctx, wrangler); err != nil {
		return err
	}

	if err := cspadaptercharts.Register(ctx, wrangler); err != nil {
		return err
	}

	clusterconnected.Register(ctx, wrangler)

	if features.MCM.Enabled() {
		hostedcluster.Register(ctx, wrangler)
	}

	if features.Fleet.Enabled() {
		if err := fleetcharts.Register(ctx, wrangler); err != nil {
			return err
		}
	}

	if features.ProvisioningV2.Enabled() || features.MCM.Enabled() {
		clusterregistrationtoken.Register(ctx, wrangler)
	}

	if features.ProvisioningV2.Enabled() {
		clusterindex.Register(ctx, wrangler)
		if err := provisioningv2.Register(ctx, wrangler); err != nil {
			return err
		}
	}

	if features.MCMAgent.Enabled() || features.MCM.Enabled() {
		err := mcmagent.Register(ctx, wrangler)
		if err != nil {
			return err
		}
	}

	return nil
}
