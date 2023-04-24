// Copyright (c) 2023, Oracle and/or its affiliates.

// This file from the Rancher repository has been modified by Oracle as follows:
// - references to the etcdsnapshots.rke.cattle.io CRDs and APIs have been removed

package provisioningv2

import (
	"context"

	"github.com/rancher/rancher/pkg/controllers/provisioningv2/cluster"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/fleetcluster"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/fleetworkspace"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/managedchart"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2/dynamicschema"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2/machinedrain"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2/machineprovision"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2/managesystemagent"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2/provisioninglog"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2/secret"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2/unmanaged"
	"github.com/rancher/rancher/pkg/features"
	"github.com/rancher/rancher/pkg/wrangler"
)

func Register(ctx context.Context, clients *wrangler.Context) error {
	cluster.Register(ctx, clients)

	if features.Fleet.Enabled() {
		managedchart.Register(ctx, clients)
		fleetcluster.Register(ctx, clients)
		fleetworkspace.Register(ctx, clients)
	}

	if features.RKE2.Enabled() {
		if features.MCM.Enabled() {
			dynamicschema.Register(ctx, clients)
			machineprovision.Register(ctx, clients)
		}
		provisioninglog.Register(ctx, clients)
		secret.Register(ctx, clients)
		unmanaged.Register(ctx, clients)
		managesystemagent.Register(ctx, clients)
		machinedrain.Register(ctx, clients)
	}

	return nil
}
