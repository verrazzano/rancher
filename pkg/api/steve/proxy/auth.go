package proxy

import (
	"net/http"
	"strings"

	"github.com/rancher/remotedialer"
	apimgmtv3 "github.com/verrazzano/rancher/pkg/apis/management.cattle.io/v3"
	v3 "github.com/verrazzano/rancher/pkg/generated/controllers/management.cattle.io/v3"
	"github.com/verrazzano/rancher/pkg/wrangler"
	apierror "k8s.io/apimachinery/pkg/api/errors"
)

const (
	tokenIndex = "clusterToken"
	Prefix     = "stv-cluster-"
)

type clusterProxyAuthorizer struct {
	tokenCache v3.ClusterRegistrationTokenCache
}

func NewAuthorizer(wrangler *wrangler.Context) remotedialer.Authorizer {
	a := &clusterProxyAuthorizer{
		tokenCache: wrangler.Mgmt.ClusterRegistrationToken().Cache(),
	}
	a.tokenCache.AddIndexer(tokenIndex, func(obj *apimgmtv3.ClusterRegistrationToken) ([]string, error) {
		return []string{obj.Status.Token}, nil
	})

	return a.Authorize
}

func (a *clusterProxyAuthorizer) Authorize(req *http.Request) (string, bool, error) {
	auth := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	if !strings.HasPrefix(auth, Prefix) {
		return "", false, nil
	}
	crts, err := a.tokenCache.GetByIndex(tokenIndex, strings.TrimPrefix(auth, Prefix))
	if apierror.IsNotFound(err) || len(crts) == 0 {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}

	return Prefix + crts[0].Namespace, true, nil
}
