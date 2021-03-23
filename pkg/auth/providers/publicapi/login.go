package publicapi

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/pkg/auth/providers"
	"github.com/rancher/rancher/pkg/auth/providers/activedirectory"
	"github.com/rancher/rancher/pkg/auth/providers/azure"
	"github.com/rancher/rancher/pkg/auth/providers/github"
	"github.com/rancher/rancher/pkg/auth/providers/googleoauth"
	"github.com/rancher/rancher/pkg/auth/providers/ldap"
	"github.com/rancher/rancher/pkg/auth/providers/local"
	"github.com/rancher/rancher/pkg/auth/providers/saml"
	"github.com/rancher/rancher/pkg/auth/tokens"
	"github.com/rancher/rancher/pkg/auth/util"
	"github.com/rancher/rancher/pkg/settings"
	v3 "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/apis/management.cattle.io/v3public"
	"github.com/rancher/types/apis/management.cattle.io/v3public/schema"
	client "github.com/rancher/types/client/management/v3public"
	"github.com/rancher/types/config"
	"github.com/rancher/types/user"
	"github.com/sirupsen/logrus"
)

const (
	CookieName = "R_SESS"
)

func newLoginHandler(ctx context.Context, mgmt *config.ScaledContext) *loginHandler {
	return &loginHandler{
		userMGR:  mgmt.UserManager,
		tokenMGR: tokens.NewManager(ctx, mgmt),
	}
}

type loginHandler struct {
	userMGR  user.Manager
	tokenMGR *tokens.Manager
}

func (h *loginHandler) login(actionName string, action *types.Action, request *types.APIContext) error {
	if actionName != "login" {
		return httperror.NewAPIError(httperror.ActionNotAvailable, "")
	}

	w := request.Response

	token, responseType, err := h.createLoginToken(request)
	if err != nil {
		// if user fails to authenticate, hide the details of the exact error. bad credentials will already be APIErrors
		// otherwise, return a generic error message
		if httperror.IsAPIError(err) {
			return err
		}
		return httperror.WrapAPIError(err, httperror.ServerError, "Server error while authenticating")
	}

	if responseType == "cookie" {
		tokenCookie := &http.Cookie{
			Name:     CookieName,
			Value:    token.ObjectMeta.Name + ":" + token.Token,
			Secure:   true,
			Path:     "/",
			HttpOnly: true,
		}
		http.SetCookie(w, tokenCookie)
	} else if responseType == "saml" {
		return nil
	} else {
		tokenData, err := tokens.ConvertTokenResource(request.Schemas.Schema(&schema.PublicVersion, client.TokenType), token)
		if err != nil {
			return httperror.WrapAPIError(err, httperror.ServerError, "Server error while authenticating")
		}
		tokenData["token"] = token.ObjectMeta.Name + ":" + token.Token
		request.WriteResponse(http.StatusCreated, tokenData)
	}

	return nil
}

func (h *loginHandler) createLoginToken(request *types.APIContext) (v3.Token, string, error) {
	var userPrincipal v3.Principal
	var groupPrincipals []v3.Principal
	var providerToken string
	logrus.Debugf("Create Token Invoked")

	bytes, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		logrus.Errorf("login failed with error: %v", err)
		return v3.Token{}, "", httperror.NewAPIError(httperror.InvalidBodyContent, "")
	}

	generic := &v3public.GenericLogin{}
	err = json.Unmarshal(bytes, generic)
	if err != nil {
		logrus.Errorf("unmarshal failed with error: %v", err)
		return v3.Token{}, "", httperror.NewAPIError(httperror.InvalidBodyContent, "")
	}
	responseType := generic.ResponseType
	description := generic.Description
	ttl := generic.TTLMillis

	authTimeout := settings.AuthUserSessionTTLMinutes.Get()
	if minutes, err := strconv.ParseInt(authTimeout, 10, 64); err == nil {
		ttl = minutes * 60 * 1000
	}

	var input interface{}
	var providerName string
	switch request.Type {
	case client.LocalProviderType:
		input = &v3public.BasicLogin{}
		providerName = local.Name
	case client.GithubProviderType:
		input = &v3public.GithubLogin{}
		providerName = github.Name
	case client.ActiveDirectoryProviderType:
		input = &v3public.BasicLogin{}
		providerName = activedirectory.Name
	case client.AzureADProviderType:
		input = &v3public.AzureADLogin{}
		providerName = azure.Name
	case client.OpenLdapProviderType:
		input = &v3public.BasicLogin{}
		providerName = ldap.OpenLdapName
	case client.FreeIpaProviderType:
		input = &v3public.BasicLogin{}
		providerName = ldap.FreeIpaName
	case client.PingProviderType:
		input = &v3public.SamlLoginInput{}
		providerName = saml.PingName
	case client.ADFSProviderType:
		input = &v3public.SamlLoginInput{}
		providerName = saml.ADFSName
	case client.KeyCloakProviderType:
		input = &v3public.SamlLoginInput{}
		providerName = saml.KeyCloakName
	case client.OKTAProviderType:
		input = &v3public.SamlLoginInput{}
		providerName = saml.OKTAName
	case client.ShibbolethProviderType:
		input = &v3public.SamlLoginInput{}
		providerName = saml.ShibbolethName
	case client.GoogleOAuthProviderType:
		input = &v3public.GoogleOauthLogin{}
		providerName = googleoauth.Name
	default:
		return v3.Token{}, "", httperror.NewAPIError(httperror.ServerError, "unknown authentication provider")
	}

	err = json.Unmarshal(bytes, input)
	if err != nil {
		logrus.Errorf("unmarshal failed with error: %v", err)
		return v3.Token{}, "", httperror.NewAPIError(httperror.InvalidBodyContent, "")
	}

	// Authenticate User
	// SAML's login flow is different from the other providers. Unlike the other providers, it gets the logged in user's data via a POST from
	// the identity provider on a separate endpoint specifically for that.

	if providerName == saml.PingName || providerName == saml.ADFSName || providerName == saml.KeyCloakName ||
		providerName == saml.OKTAName || providerName == saml.ShibbolethName {
		err = saml.PerformSamlLogin(providerName, request, input)
		return v3.Token{}, "saml", err
	}

	ctx := context.WithValue(request.Request.Context(), util.RequestKey, request.Request)
	userPrincipal, groupPrincipals, providerToken, err = providers.AuthenticateUser(ctx, input, providerName)
	if err != nil {
		return v3.Token{}, "", err
	}

	displayName := userPrincipal.DisplayName
	if displayName == "" {
		displayName = userPrincipal.LoginName
	}
	user, err := h.userMGR.EnsureUser(userPrincipal.Name, displayName)
	if err != nil {
		return v3.Token{}, "", err
	}

	if user.Enabled != nil && !*user.Enabled {
		return v3.Token{}, "", httperror.NewAPIError(httperror.PermissionDenied, "Permission Denied")
	}

	if strings.HasPrefix(responseType, tokens.KubeconfigResponseType) {
		token, err := tokens.GetKubeConfigToken(user.Name, responseType, h.userMGR)
		if err != nil {
			return v3.Token{}, "", err
		}
		return *token, responseType, nil
	}

	rToken, err := h.tokenMGR.NewLoginToken(user.Name, userPrincipal, groupPrincipals, providerToken, ttl, description)
	return rToken, responseType, err
}
