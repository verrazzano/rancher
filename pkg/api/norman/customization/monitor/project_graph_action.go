package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rancher/rancher/pkg/types/config"
	"k8s.io/apimachinery/pkg/labels"
	"net/http"
	"strings"
	"time"

	v32 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	mgmtclientv3 "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/rancher/rancher/pkg/clustermanager"
	v3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
	pv3 "github.com/rancher/rancher/pkg/generated/norman/project.cattle.io/v3"
	monitorutil "github.com/rancher/rancher/pkg/monitoring"
	"github.com/rancher/rancher/pkg/ref"
	"github.com/rancher/rancher/pkg/types/config/dialer"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewProjectGraphHandler(dialerFactory dialer.Factory, clustermanager *clustermanager.Manager) *ProjectGraphHandler {
	return &ProjectGraphHandler{
		dialerFactory:  dialerFactory,
		clustermanager: clustermanager,
		projectLister:  clustermanager.ScaledContext.Management.Projects(metav1.NamespaceAll).Controller().Lister(),
		appLister:      clustermanager.ScaledContext.Project.Apps(metav1.NamespaceAll).Controller().Lister(),
	}
}

type ProjectGraphHandler struct {
	dialerFactory  dialer.Factory
	clustermanager *clustermanager.Manager
	projectLister  v3.ProjectLister
	appLister      pv3.AppLister
}

func (h *ProjectGraphHandler) QuerySeriesAction(actionName string, action *types.Action, apiContext *types.APIContext) error {
	var queryGraphInput v32.QueryGraphInput
	actionInput, err := parse.ReadBody(apiContext.Request)
	if err != nil {
		return err
	}

	if err = convert.ToObj(actionInput, &queryGraphInput); err != nil {
		return err
	}

	inputParser := newProjectGraphInputParser(queryGraphInput)
	if err = inputParser.parse(); err != nil {
		return err
	}

	clusterName := inputParser.ClusterName
	userContext, err := h.clustermanager.UserContextNoControllers(clusterName)
	if err != nil {
		return fmt.Errorf("get usercontext failed, %v", err)
	}

	check := newAuthChecker(apiContext.Request.Context(), userContext, inputParser.Input, inputParser.ProjectID)
	if err = check.check(); err != nil {
		return err
	}

	reqContext, cancel := context.WithTimeout(context.Background(), prometheusReqTimeout)
	defer cancel()

	var svcName, svcNamespace, svcPort, token string
	var queries []*PrometheusQuery
	prometheusName, prometheusNamespace := monitorutil.ClusterMonitoringInfo()
	token, err = getAuthToken(userContext, prometheusName, prometheusNamespace)
	if err != nil {
		return err
	}

	if inputParser.Input.Filters["resourceType"] != "istioproject" {
		
		svcName, svcNamespace, svcPort = monitorutil.ClusterPrometheusEndpoint()

		var graphs []mgmtclientv3.ProjectMonitorGraph
		err = access.List(apiContext, apiContext.Version, mgmtclientv3.ProjectMonitorGraphType, &types.QueryOptions{Conditions: inputParser.Conditions}, &graphs)
		if err != nil {
			return err
		}

		mgmtClient := h.clustermanager.ScaledContext.Management
		for _, graph := range graphs {
			g := graph
			_, projectName := ref.Parse(graph.ProjectID)
			refName := getRefferenceGraphName(projectName, graph.Name)
			monitorMetrics, err := graph2Metrics(userContext, mgmtClient, clusterName, g.ResourceType, refName, graph.MetricsSelector, graph.DetailsMetricsSelector, inputParser.Input.MetricParams, inputParser.Input.IsDetails)
			if err != nil {
				return err
			}

			queries = append(queries, metrics2PrometheusQuery(monitorMetrics, inputParser.Start, inputParser.End, inputParser.Step, isInstanceGraph(g.GraphType))...)
		}
	}

	prometheusQuery, err := NewPrometheusQuery(reqContext, clusterName, token, svcNamespace, svcName, svcPort, h.dialerFactory, userContext)
	if err != nil {
		return err
	}
	seriesSlice, err := prometheusQuery.Do(queries)
	if err != nil {
		logrus.WithError(err).Warn("query series failed")
		return httperror.NewAPIError(httperror.ServerError, "Failed to obtain metrics. The metrics service may not be available.")
	}

	if seriesSlice == nil {
		apiContext.WriteResponse(http.StatusNoContent, nil)
		return nil
	}

	collection := v32.QueryProjectGraphOutput{Type: "collection"}
	for k, v := range seriesSlice {
		graphName, _, _ := parseID(k)
		queryGraph := v32.QueryProjectGraph{
			GraphName: graphName,
			Series:    parseResponse(v),
		}
		collection.Data = append(collection.Data, queryGraph)
	}

	res, err := json.Marshal(collection)
	if err != nil {
		return fmt.Errorf("marshal query series result failed, %v", err)
	}
	apiContext.Response.Write(res)
	return nil
}

func parseResponse(seriesSlice []*TimeSeries) []*v32.TimeSeries {
	var series []*v32.TimeSeries
	for _, v := range seriesSlice {
		series = append(series, &v32.TimeSeries{
			Name:   v.Name,
			Points: v.Points,
		})
	}
	return series
}

func getRefferenceGraphName(namespace, name string) string {
	return fmt.Sprintf("%s:%s", namespace, name)
}

func graph2Metrics(userContext *config.UserContext, mgmtClient v3.Interface, clusterName, resourceType, refGraphName string, metricSelector, detailsMetricSelector map[string]string, metricParams map[string]string, isDetails bool) ([]*metricWrap, error) {
	projectName, _ := ref.Parse(refGraphName)
	nodeLister := mgmtClient.Nodes(metav1.NamespaceAll).Controller().Lister()
	newMetricParams, err := parseMetricParams(userContext, nodeLister, resourceType, clusterName, projectName, metricParams)
	if err != nil {
		return nil, err
	}

	var excuteMetrics []*metricWrap
	var set labels.Set
	if isDetails && detailsMetricSelector != nil {
		set = labels.Set(detailsMetricSelector)
	} else {
		set = labels.Set(metricSelector)
	}
	metrics, err := mgmtClient.MonitorMetrics(clusterName).List(metav1.ListOptions{LabelSelector: set.AsSelector().String()})
	if err != nil {
		return nil, fmt.Errorf("list metrics failed, %v", err)
	}

	for _, v := range metrics.Items {
		executeExpression := replaceParams(newMetricParams, v.Spec.Expression)
		excuteMetrics = append(excuteMetrics, &metricWrap{
			MonitorMetric:              *v.DeepCopy(),
			ExecuteExpression:          executeExpression,
			ReferenceGraphName:         refGraphName,
			ReferenceGraphResourceType: resourceType,
		})
	}
	return excuteMetrics, nil
}

func metrics2PrometheusQuery(metrics []*metricWrap, start, end time.Time, step time.Duration, isInstanceQuery bool) []*PrometheusQuery {
	var queries []*PrometheusQuery
	for _, v := range metrics {
		id := getPrometheusQueryID(v.ReferenceGraphName, v.ReferenceGraphResourceType, v.Name)
		queries = append(queries, InitPromQuery(id, start, end, step, v.ExecuteExpression, v.Spec.LegendFormat, isInstanceQuery))
	}
	return queries
}

func getPrometheusQueryID(graphName, graphResoureceType, metricName string) string {
	return fmt.Sprintf("%s_%s_%s", graphName, graphResoureceType, metricName)
}

func parseID(ref string) (graphName, resourceType, metricName string) {
	parts := strings.SplitN(ref, "_", 3)

	if len(parts) < 2 {
		return parts[0], "", ""
	}

	if len(parts) == 2 {
		return parts[0], parts[1], ""
	}

	if len(parts) == 3 {
		return parts[0], parts[1], parts[2]
	}
	return parts[0], parts[1], parts[1]
}

type metricWrap struct {
	v3.MonitorMetric
	ExecuteExpression          string
	ReferenceGraphName         string
	ReferenceGraphResourceType string
}
