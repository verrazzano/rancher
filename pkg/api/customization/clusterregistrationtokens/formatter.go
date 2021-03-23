package clusterregistrationtokens

import (
	"fmt"
	"net/url"

	"github.com/rancher/norman/types"
	util "github.com/rancher/rancher/pkg/cluster"
	"github.com/rancher/rancher/pkg/image"
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rancher/pkg/systemtemplate"
	v3 "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	commandFormat            = "kubectl apply -f %s"
	insecureCommandFormat    = "curl --insecure -sfL %s | kubectl apply -f -"
	nodeCommandFormat        = "sudo docker run -d --privileged --restart=unless-stopped --net=host -v /etc/kubernetes:/etc/kubernetes -v /var/run:/var/run %s --server %s --token %s%s"
	loginCommandFormat       = "echo \"%s\" | sudo docker login --username %s --password-stdin %s"
	windowsNodeCommandFormat = `PowerShell -NoLogo -NonInteractive -Command "& {docker run -v c:\:c:\host %s%s bootstrap --server %s --token %s%s%s | iex}"`
)

type Formatter struct {
	Clusters v3.ClusterInterface
}

func NewFormatter(managementContext *config.ScaledContext) *Formatter {
	clusterFormatter := Formatter{
		Clusters: managementContext.Management.Clusters(""),
	}
	return &clusterFormatter
}

func (f *Formatter) Formatter(request *types.APIContext, resource *types.RawResource) {
	ca := systemtemplate.CAChecksum()
	if ca != "" {
		ca = " --ca-checksum " + ca
	}

	token, _ := resource.Values["token"].(string)
	if token != "" {
		url := getURL(request, token)
		resource.Values["insecureCommand"] = fmt.Sprintf(insecureCommandFormat, url)
		resource.Values["command"] = fmt.Sprintf(commandFormat, url)
		resource.Values["token"] = token
		resource.Values["manifestUrl"] = url
		rootURL := getRootURL(request)

		cluster, _ := f.Clusters.Get(fmt.Sprintf("%v", resource.Values["clusterId"]), metav1.GetOptions{})

		agentImage := image.ResolveWithCluster(settings.AgentImage.Get(), cluster)
		// for linux
		resource.Values["nodeCommand"] = fmt.Sprintf(nodeCommandFormat,
			agentImage,
			rootURL,
			token,
			ca)
		// for windows
		var agentImageDockerEnv string
		if util.GetPrivateRepoURL(cluster) != "" {
			// patch the AGENT_IMAGE env
			agentImageDockerEnv = fmt.Sprintf("-e AGENT_IMAGE=%s ", agentImage)
		}

		resource.Values["windowsNodeCommand"] = fmt.Sprintf(windowsNodeCommandFormat,
			agentImageDockerEnv,
			agentImage,
			rootURL,
			token,
			ca,
			getWindowsPrefixPathArg(cluster.Spec.RancherKubernetesEngineConfig))
	}
}

func getWindowsPrefixPathArg(rkeConfig *v3.RancherKubernetesEngineConfig) string {
	if rkeConfig == nil {
		return ""
	}
	// default to prefix path
	prefixPath := rkeConfig.PrefixPath

	// if windows prefix path set, override
	if rkeConfig.WindowsPrefixPath != "" {
		prefixPath = rkeConfig.WindowsPrefixPath
	}

	if prefixPath != "" {
		return fmt.Sprintf(" --prefix-path %s", prefixPath)
	}

	return ""
}

func NodeCommand(token string, cluster *v3.Cluster) string {
	ca := systemtemplate.CAChecksum()
	if ca != "" {
		ca = " --ca-checksum " + ca
	}

	return fmt.Sprintf(nodeCommandFormat,
		image.ResolveWithCluster(settings.AgentImage.Get(), cluster),
		getRootURL(nil),
		token,
		ca)
}

func LoginCommand(reg v3.PrivateRegistry) string {
	return fmt.Sprintf(
		loginCommandFormat,
		// escape password special characters so it is interpreted correctly when command is executed
		escapeSpecialChars(reg.Password),
		reg.User,
		reg.URL,
	)
}

// escapeSpecialChars escapes ", `, $, \ from a string s
func escapeSpecialChars(s string) string {
	var escaped []rune
	for _, r := range s {
		switch r {
		case '"', '`', '$', '\\': // escape
			escaped = append(escaped, '\\', r)
		default: // no escape
			escaped = append(escaped, r)
		}
	}
	return string(escaped)
}

func getRootURL(request *types.APIContext) string {
	serverURL := settings.ServerURL.Get()
	if serverURL == "" {
		if request != nil {
			serverURL = request.URLBuilder.RelativeToRoot("")
		}
	} else {
		u, err := url.Parse(serverURL)
		if err != nil {
			if request != nil {
				serverURL = request.URLBuilder.RelativeToRoot("")
			}
		} else {
			u.Path = ""
			serverURL = u.String()
		}
	}

	return serverURL
}

func getURL(request *types.APIContext, token string) string {
	path := "/v3/import/" + token + ".yaml"
	serverURL := settings.ServerURL.Get()
	if serverURL == "" {
		serverURL = request.URLBuilder.RelativeToRoot(path)
	} else {
		u, err := url.Parse(serverURL)
		if err != nil {
			serverURL = request.URLBuilder.RelativeToRoot(path)
		} else {
			u.Path = path
			serverURL = u.String()
		}
	}

	return serverURL
}
