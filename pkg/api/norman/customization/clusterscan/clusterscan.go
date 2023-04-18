package clusterscan

import (
	"net/http"
	"strconv"

	v32 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/pkg/clustermanager"
	corev1 "github.com/rancher/rancher/pkg/generated/norman/core/v1"
	"github.com/rancher/rancher/pkg/ref"
	"github.com/rancher/security-scan/pkg/kb-summarizer/report"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	errorState  = "error"
	failedState = "fail"
	passedState = "pass"
)

type Handler struct {
	CoreClient     corev1.Interface
	ClusterManager *clustermanager.Manager
}

func (h Handler) LinkHandler(apiContext *types.APIContext, next types.RequestHandler) error {
	var cs map[string]interface{}
	if err := access.ByID(apiContext, apiContext.Version, apiContext.Type, apiContext.ID, &cs); err != nil {
		return err
	}

	clusterID, clusterScanID := ref.Parse(cs["id"].(string))

	clusterContext, err := h.ClusterManager.UserContextNoControllers(clusterID)
	if err != nil {
		return err
	}

	cm, err := clusterContext.Core.ConfigMaps(v32.DefaultNamespaceForCis).Get(clusterScanID, metav1.GetOptions{})
	if err != nil {
		return err
	}

	reportJSON, err := report.GetJSONBytes([]byte(cm.Data[v32.DefaultScanOutputFileName]))
	if err != nil {
		return err
	}

	apiContext.Response.Header().Set("Content-Length", strconv.Itoa(len(reportJSON)))
	apiContext.Response.Header().Set("Content-Type", "application/json")
	apiContext.Response.WriteHeader(http.StatusOK)
	_, err = apiContext.Response.Write(reportJSON)
	if err != nil {
		return err
	}

	return nil
}
