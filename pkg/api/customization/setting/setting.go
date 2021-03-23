package setting

import (
	"fmt"
	"strings"
	"time"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/slice"
	"github.com/rancher/rancher/pkg/auth/providerrefresh"
	"github.com/rancher/rancher/pkg/auth/tokens"
	"github.com/rancher/rancher/pkg/settings"
	v3 "github.com/rancher/types/apis/management.cattle.io/v3"
	v3client "github.com/rancher/types/client/management/v3"
)

var ReadOnlySettings = []string{
	"cacerts",
}

func PipelineFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	v, ok := resource.Values["value"]
	if !ok || v == "" {
		resource.Values["value"] = resource.Values["default"]
		resource.Values["customized"] = false
	} else {
		resource.Values["customized"] = true
	}
}

func Formatter(apiContext *types.APIContext, resource *types.RawResource) {
	if convert.ToString(resource.Values["source"]) == "env" {
		delete(resource.Links, "update")
	} else if slice.ContainsString(ReadOnlySettings, resource.ID) {
		delete(resource.Links, "update")
	} else {
		setting := map[string]interface{}{
			"id": apiContext.ID,
		}
		if err := apiContext.AccessControl.CanDo(v3.SettingGroupVersionKind.Group, v3.SettingResource.Name, "update", apiContext, setting, apiContext.Schema); err != nil {
			delete(resource.Links, "update")
		}
	}
}

func Validator(request *types.APIContext, schema *types.Schema, data map[string]interface{}) error {
	var setting v3client.Setting

	// request.ID is taken from the request request url, it is possible that the request url does not contain the id
	id := request.ID
	if name, ok := data["name"].(string); ok && id == "" {
		id = name
	}

	if err := access.ByID(request, request.Version, v3client.SettingType, id, &setting); err != nil {
		if !httperror.IsNotFound(err) {
			return err
		}
	}
	if setting.Source == "env" {
		return httperror.NewAPIError(httperror.MethodNotAllowed, fmt.Sprintf("%s is readOnly because its value is from environment variable", id))
	} else if slice.ContainsString(ReadOnlySettings, id) {
		return httperror.NewAPIError(httperror.MethodNotAllowed, fmt.Sprintf("%s is readOnly", id))
	}

	newValue, ok := data["value"]
	if !ok {
		return fmt.Errorf("value not found")
	}
	newValueString, ok := newValue.(string)
	if !ok {
		return fmt.Errorf("value not string")
	}

	var err error
	switch id {
	case "auth-user-info-max-age-seconds":
		_, err = providerrefresh.ParseMaxAge(newValueString)
	case "auth-user-info-resync-cron":
		_, err = providerrefresh.ParseCron(newValueString)
	case "kubeconfig-token-ttl-minutes":
		generateToken := strings.EqualFold(settings.KubeconfigGenerateToken.Get(), "true")
		if generateToken {
			return httperror.NewAPIError(httperror.ActionNotAvailable, fmt.Sprintf("kubeconfig-token-ttl-minutes can be set only if rancher doesn't generate token, "+
				"disable kubeconfig-generate-token"))
		}

		var tokenTTL time.Duration
		tokenTTL, err = tokens.ParseTokenTTL(newValueString)
		if err == nil {
			maxTTL, err := tokens.ParseTokenTTL(settings.AuthTokenMaxTTLMinutes.Get())
			if err != nil {
				return httperror.NewAPIError(httperror.InvalidBodyContent,
					fmt.Sprintf("error parsing auth-token-max-ttl-minutes %v", err))
			}
			if maxTTL != 0 {
				if tokenTTL == 0 || tokenTTL.Minutes() > maxTTL.Minutes() {
					return httperror.NewAPIError(httperror.MaxLimitExceeded,
						fmt.Sprintf("max ttl for tokens is [%s]", settings.AuthTokenMaxTTLMinutes.Get()))
				}
			}
		}
	}

	if err != nil {
		return httperror.NewAPIError(httperror.InvalidBodyContent, fmt.Sprintf("%v", err))
	}

	return nil
}
