package mcmagent

import (
	"context"

	"github.com/verrazzano/rancher/pkg/controllers/managementagent"
	"github.com/verrazzano/rancher/pkg/types/config"
	"github.com/verrazzano/rancher/pkg/wrangler"
)

func Register(ctx context.Context, wrangler *wrangler.Context) error {
	userContext, err := config.NewUserOnlyContext(wrangler)
	if err != nil {
		return err
	}
	return managementagent.Register(ctx, userContext)
}
