package image

import (
	"fmt"
	"path"
	"strings"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	v1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/rancher/pkg/settings"
)

func ResolveWithControlPlane(image string, cp *rkev1.RKEControlPlane) string {
	return resolve(GetPrivateRepoURLFromControlPlane(cp), image)
}

func ResolveWithCluster(image string, cluster *v1.Cluster) string {
	return resolve(GetPrivateRepoURLFromCluster(cluster), image)
}

func resolve(reg, image string) string {
	fmt.Println(fmt.Sprintf("provisioning2.resolve: reg: %s, image: %s", reg, image))
	if reg != "" && !strings.HasPrefix(image, reg) {
		////Images from Dockerhub Library repo, we add rancher prefix when using private registry
		//if !strings.Contains(image, "/") {
		//	image = "rancher/" + image
		//}
		//return path.Join(reg, image)
		/*
			Separating the image from the default registry url through split
			Ex:
			input:
			reg (private registry url) -> myreg.io/myrepo
			image ->  ghcr.io/verrazzano/rancher-agent:1.2.3

			imageSplit := strings.SplitN(image, "/", 2)

			output: imageSplit -> [ghcr.io verrazzano/rancher-agent:1.2.3]
			path.Join(reg, imageSplit[1]) -> myreg.io/myrepo/verrazzano/rancher-agent:1.2.3
			Therefore, we only use the imageSplit[1] -> verrazzano/rancher-agent:1.2.3
			for concatenation with the reg.
		*/
		imageSplit := strings.SplitN(image, "/", 2)
		//Concatenating only the image name with the private registry url.
		imagePath := path.Join(reg, imageSplit[1])
		fmt.Println(fmt.Sprintf("provisioning2.resolve: imagePath: %s", imagePath))
		return imagePath
	}

	return image
}

// GetPrivateRepoSecretFromCluster returns the name of the secret containing the credentials for the cluster level system-default-registry.
func GetPrivateRepoSecretFromCluster(cluster *v1.Cluster) string {
	if cluster != nil && cluster.Spec.RKEConfig != nil && cluster.Spec.RKEConfig.Registries != nil {
		return cluster.Spec.RKEConfig.Registries.Configs[GetPrivateRepoURLFromCluster(cluster)].AuthConfigSecretName
	}
	return "" // don't return settings.SystemDefaultRegistry.Get() as the global system-default-registry requires no auth
}

// GetPrivateRepoURLFromCluster returns the system-default-registry URL from either the clusters
// machineGlobalConfig, or one of its machineSelectorConfig's which has no label selectors.
// If no cluster level system-default-registry is configured, it will return the global system-default-registry.
func GetPrivateRepoURLFromCluster(cluster *v1.Cluster) string {
	if cluster != nil && cluster.Spec.RKEConfig != nil {
		return getPrivateRepoURL(cluster.Spec.RKEConfig.MachineGlobalConfig, cluster.Spec.RKEConfig.MachineSelectorConfig)
	}

	return settings.SystemDefaultRegistry.Get()
}

// GetPrivateRepoURLFromControlPlane returns the system-default-registry URL from either the control planes
// machineGlobalConfig, or one of its machineSelectorConfig's which has no label selectors.
// If no cluster level system-default-registry is configured, it will return the global system-default-registry.
func GetPrivateRepoURLFromControlPlane(cp *rkev1.RKEControlPlane) string {
	if cp != nil {
		return getPrivateRepoURL(cp.Spec.MachineGlobalConfig, cp.Spec.MachineSelectorConfig)
	}

	return settings.SystemDefaultRegistry.Get()
}

func getPrivateRepoURL(machineGlobalConfig rkev1.GenericMap, machineSelectorConfig []rkev1.RKESystemConfig) string {
	for key, val := range machineGlobalConfig.Data {
		if val, ok := val.(string); ok && key == "system-default-registry" {
			return val
		}
	}

	for _, config := range machineSelectorConfig {
		if registryVal, ok := config.Config.Data["system-default-registry"]; config.MachineLabelSelector == nil && ok {
			if registry, ok := registryVal.(string); ok {
				return registry
			}
		}
	}

	return settings.SystemDefaultRegistry.Get()
}

func GetDesiredAgentImage(cp *rkev1.RKEControlPlane, mgmtCluster *v3.Cluster) string {
	desiredAgent := mgmtCluster.Spec.DesiredAgentImage
	if mgmtCluster.Spec.AgentImageOverride != "" {
		desiredAgent = mgmtCluster.Spec.AgentImageOverride
	}
	if desiredAgent == "" || desiredAgent == "fixed" {
		desiredAgent = ResolveWithControlPlane(settings.AgentImage.Get(), cp)
	}
	return desiredAgent
}

func GetDesiredAuthImage(cp *rkev1.RKEControlPlane, mgmtCluster *v3.Cluster) string {
	var desiredAuth string
	if mgmtCluster.Spec.LocalClusterAuthEndpoint.Enabled {
		desiredAuth = mgmtCluster.Spec.DesiredAuthImage
		if desiredAuth == "" || desiredAuth == "fixed" {
			desiredAuth = ResolveWithControlPlane(settings.AuthImage.Get(), cp)
		}
	}
	return desiredAuth
}
