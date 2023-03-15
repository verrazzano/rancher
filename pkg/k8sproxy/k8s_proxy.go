package k8sproxy

import (
	"net/http"

	"github.com/verrazzano/rancher/pkg/clusterrouter"
	"github.com/verrazzano/rancher/pkg/clusterrouter/proxy"
	"github.com/verrazzano/rancher/pkg/k8slookup"
	"github.com/verrazzano/rancher/pkg/types/config"
	"github.com/verrazzano/rancher/pkg/types/config/dialer"
)

func New(scaledContext *config.ScaledContext, dialer dialer.Factory, clusterContextGetter proxy.ClusterContextGetter) http.Handler {
	return clusterrouter.New(&scaledContext.RESTConfig, k8slookup.New(scaledContext, true), dialer,
		scaledContext.Management.Clusters("").Controller().Lister(),
		clusterContextGetter)
}
