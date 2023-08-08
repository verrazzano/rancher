package fleetcharts

import (
	"context"
	"os"
	"sync"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/controllers/dashboard/chart"
	"github.com/rancher/rancher/pkg/features"
	fleetconst "github.com/rancher/rancher/pkg/fleet"
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rancher/pkg/wrangler"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"github.com/sirupsen/logrus"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	fleetCRDChart = chart.Definition{
		ReleaseNamespace: fleetconst.ReleaseNamespace,
		ChartName:        fleetconst.CRDChartName,
	}
	fleetChart = chart.Definition{
		ReleaseNamespace: fleetconst.ReleaseNamespace,
		ChartName:        fleetconst.ChartName,
	}
	fleetUninstallChart = chart.Definition{
		ReleaseNamespace: fleetconst.ReleaseLegacyNamespace,
		ChartName:        fleetconst.ChartName,
	}
)

func Register(ctx context.Context, wContext *wrangler.Context) error {
	h := &handler{
		manager:      wContext.SystemChartsManager,
		chartsConfig: chart.RancherConfigGetter{ConfigCache: wContext.Core.ConfigMap().Cache()},
	}

	wContext.Mgmt.Setting().OnChange(ctx, "fleet-install", h.onSetting)
	// watch cluster repo `rancher-charts` and enqueue the setting to make sure the latest fleet is installed after catalog refresh
	relatedresource.WatchClusterScoped(ctx, "bootstrap-fleet-charts", func(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
		if name == "rancher-charts" {
			return []relatedresource.Key{{
				Name: settings.ServerURL.Name,
			}}, nil
		}
		return nil, nil
	}, wContext.Mgmt.Setting(), wContext.Catalog.ClusterRepo())
	return nil
}

type handler struct {
	sync.Mutex
	manager      chart.Manager
	chartsConfig chart.RancherConfigGetter
}

func (h *handler) onSetting(key string, setting *v3.Setting) (*v3.Setting, error) {
	if setting == nil {
		return nil, nil
	}

	if setting.Name != settings.ServerURL.Name &&
		setting.Name != settings.CACerts.Name &&
		setting.Name != settings.SystemDefaultRegistry.Name {
		return setting, nil
	}

	h.Lock()
	if err := h.manager.Uninstall(fleetUninstallChart.ReleaseNamespace, fleetUninstallChart.ChartName); err != nil {
		h.Unlock()
		return nil, err
	}
	h.Unlock()

	err := h.manager.Ensure(fleetCRDChart.ReleaseNamespace, fleetCRDChart.ChartName, settings.FleetMinVersion.Get(), "", nil, true, "")
	if err != nil {
		return setting, err
	}

	systemGlobalRegistry := map[string]interface{}{
		"cattle": map[string]interface{}{
			"systemDefaultRegistry": settings.SystemDefaultRegistry.Get(),
		},
	}

	fleetChartValues := map[string]interface{}{
		"apiServerURL": settings.ServerURL.Get(),
		"apiServerCA":  settings.CACerts.Get(),
		"global":       systemGlobalRegistry,
		"bootstrap": map[string]interface{}{
			"enabled":        false,
			"agentNamespace": fleetconst.ReleaseLocalNamespace,
		},
		"gitops": map[string]interface{}{
			"enabled": features.Gitops.Enabled(),
		},
	}

	overrideFleetImages(fleetChartValues)

	gitjobChartValues := make(map[string]interface{})

	if envVal, ok := os.LookupEnv("HTTP_PROXY"); ok {
		fleetChartValues["proxy"] = envVal
		gitjobChartValues["proxy"] = envVal
	}
	if envVal, ok := os.LookupEnv("NO_PROXY"); ok {
		fleetChartValues["noProxy"] = envVal
		gitjobChartValues["noProxy"] = envVal
	}

	// add priority class value
	if priorityClassName, err := h.chartsConfig.GetPriorityClassName(); err != nil {
		if !apierror.IsNotFound(err) {
			logrus.Warnf("Failed to get rancher priorityClassName for '%s': %v", fleetChart.ChartName, err)
		}
	} else {
		fleetChartValues[chart.PriorityClassKey] = priorityClassName
		gitjobChartValues[chart.PriorityClassKey] = priorityClassName
	}

	overrideGitJobImage(gitjobChartValues)

	if len(gitjobChartValues) > 0 {
		fleetChartValues["gitjob"] = gitjobChartValues
	}

	return setting, h.manager.Ensure(fleetChart.ReleaseNamespace, fleetChart.ChartName, settings.FleetMinVersion.Get(), "", fleetChartValues, true, "")
}

// overrideFleetImages overrides the Fleet image names and/or tags
func overrideFleetImages(fleetChartValues map[string]interface{}) {
	overrideFleetImage(fleetChartValues)
	overrideFleetAgentImage(fleetChartValues)
}

// overrideFleetImage sets Helm chart values to override the Fleet image name and/or tag based on environment variables.
func overrideFleetImage(fleetChartValues map[string]interface{}) {
	chartValues := make(map[string]interface{})

	if envVal, ok := os.LookupEnv("FLEET_IMAGE"); ok {
		chartValues["repository"] = envVal
	}
	if envVal, ok := os.LookupEnv("FLEET_IMAGE_TAG"); ok {
		chartValues["tag"] = envVal
	}

	if len(chartValues) > 0 {
		fleetChartValues["image"] = chartValues
	}
}

// overrideFleetAgentImage sets Helm chart values to override the Fleet Agent image name and/or tag based on environment variables.
func overrideFleetAgentImage(fleetChartValues map[string]interface{}) {
	chartValues := make(map[string]interface{})

	if envVal, ok := os.LookupEnv("FLEET_AGENT_IMAGE"); ok {
		chartValues["repository"] = envVal
	}
	if envVal, ok := os.LookupEnv("FLEET_AGENT_IMAGE_TAG"); ok {
		chartValues["tag"] = envVal
	}

	if len(chartValues) > 0 {
		fleetChartValues["agentImage"] = chartValues
	}
}

// overrideGitJobImage sets Helm chart values to override the GitJob image name and/or tag based on environment variables.
func overrideGitJobImage(gitjobChartValues map[string]interface{}) {
	chartValues := make(map[string]interface{})

	if envVal, ok := os.LookupEnv("GITJOB_IMAGE"); ok {
		chartValues["repository"] = envVal
	}
	if envVal, ok := os.LookupEnv("GITJOB_IMAGE_TAG"); ok {
		chartValues["tag"] = envVal
	}

	if len(chartValues) > 0 {
		gitjobChartValues["gitjob"] = chartValues
	}
}
