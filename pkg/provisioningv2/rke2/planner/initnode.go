package planner

import (
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2"
	"github.com/rancher/wrangler/pkg/generic"
)

// clearInitNodeMark removes the init node label on the given machine and updates the machine directly against the api
// server, effectively immediately demoting it from being an init node
func (p *Planner) clearInitNodeMark(entry *planEntry) error {
	if entry.Metadata.Labels[rke2.InitNodeLabel] == "" {
		return nil
	}

	if err := p.store.removePlanSecretLabel(entry, rke2.InitNodeLabel); err != nil {
		return err
	}
	// We've changed state, so let the caches sync up again
	return generic.ErrSkip
}

// setInitNodeMark sets the init node label on the given machine and updates the machine directly against the api
// server. It returns the modified/updated machine object
func (p *Planner) setInitNodeMark(entry *planEntry) error {
	if entry.Metadata.Labels[rke2.InitNodeLabel] == "true" {
		return nil
	}

	entry.Metadata.Labels[rke2.InitNodeLabel] = "true"
	if err := p.store.updatePlanSecretLabelsAndAnnotations(entry); err != nil {
		return err
	}

	// We've changed state, so let the caches sync up again
	return generic.ErrSkip
}
