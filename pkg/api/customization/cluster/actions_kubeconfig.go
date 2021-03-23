package cluster

import (
	"net/http"
	"strings"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/pkg/kubeconfig"
	"github.com/rancher/rancher/pkg/settings"
	mgmtclient "github.com/rancher/types/client/management/v3"
)

func (a ActionHandler) GenerateKubeconfigActionHandler(actionName string, action *types.Action, apiContext *types.APIContext) error {
	var cluster mgmtclient.Cluster
	if err := access.ByID(apiContext, apiContext.Version, apiContext.Type, apiContext.ID, &cluster); err != nil {
		return err
	}

	var (
		cfg   string
		token string
		err   error
	)

	endpointEnabled := cluster.LocalClusterAuthEndpoint != nil && cluster.LocalClusterAuthEndpoint.Enabled

	generateToken := strings.EqualFold(settings.KubeconfigGenerateToken.Get(), "true")
	if generateToken {
		// generate token and place it in kubeconfig, token doesn't expire
		if endpointEnabled {
			token, err = a.getClusterToken(cluster.ID, apiContext)
		} else {
			token, err = a.getToken(apiContext)
		}
		if err != nil {
			return err
		}
	}

	if endpointEnabled {
		cfg, err = kubeconfig.ForClusterTokenBased(&cluster, apiContext.ID, apiContext.Request.Host, token)
	} else {
		cfg, err = kubeconfig.ForTokenBased(cluster.Name, apiContext.ID, apiContext.Request.Host, token)
	}
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"config": cfg,
		"type":   "generateKubeconfigOutput",
	}
	apiContext.WriteResponse(http.StatusOK, data)
	return nil
}
