package steve

import (
	"context"
	"net/http"

	gmux "github.com/gorilla/mux"
	steve "github.com/rancher/steve/pkg/server"
	"github.com/verrazzano/rancher/pkg/api/steve/aggregation"
	"github.com/verrazzano/rancher/pkg/api/steve/github"
	"github.com/verrazzano/rancher/pkg/api/steve/health"
	"github.com/verrazzano/rancher/pkg/api/steve/projects"
	"github.com/verrazzano/rancher/pkg/api/steve/proxy"
	"github.com/verrazzano/rancher/pkg/features"
	"github.com/verrazzano/rancher/pkg/provisioningv2/rke2/configserver"
	"github.com/verrazzano/rancher/pkg/provisioningv2/rke2/installer"
	"github.com/verrazzano/rancher/pkg/settings"
	"github.com/verrazzano/rancher/pkg/wrangler"
)

func AdditionalAPIsPreMCM(config *wrangler.Context) func(http.Handler) http.Handler {
	if features.RKE2.Enabled() {
		connectHandler := configserver.New(config)
		mux := gmux.NewRouter()
		mux.UseEncodedPath()
		mux.Handle(configserver.ConnectAgent, connectHandler)
		mux.Handle(configserver.ConnectConfigYamlPath, connectHandler)
		mux.Handle(configserver.ConnectClusterInfo, connectHandler)
		mux.Handle(installer.SystemAgentInstallPath, installer.Handler)
		mux.Handle(installer.WindowsRke2InstallPath, installer.Handler)
		return func(next http.Handler) http.Handler {
			mux.NotFoundHandler = next
			return mux
		}
	}

	return func(next http.Handler) http.Handler {
		return next
	}
}

func AdditionalAPIs(ctx context.Context, config *wrangler.Context, steve *steve.Server) (func(http.Handler) http.Handler, error) {
	clusterAPI, err := projects.Projects(ctx, steve)
	if err != nil {
		return nil, err
	}

	githubHandler, err := github.NewProxy(config.Core.Secret().Cache(),
		settings.GithubProxyAPIURL.Get(),
		"cattle-system",
		"github")
	if err != nil {
		return nil, err
	}

	mux := gmux.NewRouter()
	mux.UseEncodedPath()
	mux.Handle("/v1/github{path:.*}", githubHandler)
	mux.Handle("/v3/connect", Tunnel(config))
	health.Register(mux)

	return func(next http.Handler) http.Handler {
		mux.NotFoundHandler = clusterAPI(next)
		return mux
	}, nil
}

func Tunnel(config *wrangler.Context) http.Handler {
	config.TunnelAuthorizer.Add(proxy.NewAuthorizer(config))
	config.TunnelAuthorizer.Add(aggregation.New(config))
	return config.TunnelServer
}
