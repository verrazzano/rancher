// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/oracle/oci-go-sdk/v53/common"
	"github.com/oracle/oci-go-sdk/v53/core"
	"github.com/rancher/norman/httperror"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"strings"
)

type OCNEConfig struct {
	ApplyYamls            []string `json:"applyYamls"`
	CloudCredentialID     string   `json:"cloudCredentialId"`
	ClusterCidr           string   `json:"clusterCidr"`
	ClusterName           string   `json:"clusterName"`
	CnePath               string   `json:"cnePath"`
	CompartmentID         string   `json:"compartmentId"`
	ControlPlaneMemoryGbs int      `json:"controlPlaneMemoryGbs"`
	ControlPlaneOcpus     int      `json:"controlPlaneOcpus"`
	ControlPlaneShape     string   `json:"controlPlaneShape"`
	ControlPlaneSubnet    string   `json:"controlPlaneSubnet"`
	ControlPlaneVolumeGbs int      `json:"controlPlaneVolumeGbs"`
	CorednsImageTag       string   `json:"corednsImageTag"`
	DisplayName           string   `json:"displayName"`
	DriverName            string   `json:"driverName"`
	EtcdImageTag          string   `json:"etcdImageTag"`
	ImageDisplayName      string   `json:"imageDisplayName"`
	ImageID               string   `json:"imageId"`
	InstallCalico         bool     `json:"installCalico"`
	InstallCcm            bool     `json:"installCcm"`
	InstallVerrazzano     bool     `json:"installVerrazzano"`
	KubernetesVersion     string   `json:"kubernetesVersion"`
	LoadBalancerSubnet    string   `json:"loadBalancerSubnet"`
	Name                  string   `json:"name"`
	NodePools             []string `json:"nodePools"`
	NodePublicKeyContents string   `json:"nodePublicKeyContents"`
	NodeShape             string   `json:"nodeShape"`
	NumControlPlaneNodes  int      `json:"numControlPlaneNodes"`
	NumWorkerNodes        int      `json:"numWorkerNodes"`
	OcneVersion           string   `json:"ocneVersion"`
	PodCidr               string   `json:"podCidr"`
	PrivateRegistry       string   `json:"privateRegistry"`
	ProxyEndpoint         string   `json:"proxyEndpoint"`
	Region                string   `json:"region"`
	SkipOcneInstall       bool     `json:"skipOcneInstall"`
	TigeraImageTag        string   `json:"tigeraImageTag"`
	Type                  string   `json:"type"`
	UseNodePvEncryption   bool     `json:"useNodePvEncryption"`
	VcnID                 string   `json:"vcnId"`
	VerrazzanoResource    string   `json:"verrazzanoResource"`
	VerrazzanoTag         string   `json:"verrazzanoTag"`
	VerrazzanoVersion     string   `json:"verrazzanoVersion"`
	WorkerNodeSubnet      string   `json:"workerNodeSubnet"`

	net *core.VirtualNetworkClient
}

type NodePool struct {
	Name       string `json:"name"`
	Replicas   int64  `json:"replicas"`
	Memory     int64  `json:"memory"`
	Ocpus      int64  `json:"ocpus"`
	VolumeSize int64  `json:"volumeSize"`
	Shape      string `json:"shape"`
}

var BadRequest = httperror.ErrorCode{
	Code:   "Bad Request",
	Status: 400,
}

func (v *Validator) ValidateOCNE(ocneConfig map[string]interface{}) error {
	o, err := getOCNEConfig(ocneConfig)
	if err != nil {
		return err
	}
	if err := v.validateOCNECredentials(o); err != nil {
		return err
	}

	errorChannel := make(chan error)
	validators := []func(){
		func() {
			v.validateVerrazzano(errorChannel, o)
		},
		func() {
			v.validateOCNECluster(errorChannel, o)
		},
		func() {
			v.validateAdditionalYAMLs(errorChannel, o)
		},
	}

	for _, validatorFunc := range validators {
		go validatorFunc()
	}

	var result error
	for i := 0; i < len(validators); i++ {
		err = <-errorChannel
		if err != nil {
			result = err
		}
	}
	close(errorChannel)
	return result
}

func getOCNEConfig(ocneConfig map[string]interface{}) (*OCNEConfig, error) {
	b, err := json.Marshal(ocneConfig)
	if err != nil {
		return nil, err
	}
	o := &OCNEConfig{}
	if err := json.Unmarshal(b, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (v *Validator) validateOCNECredentials(o *OCNEConfig) error {
	if o.CloudCredentialID == "" {
		return httperror.NewAPIError(BadRequest, `"cloudCredentialId" field is required.`)
	}

	// CloudCredentialID is generally in the form <namespace>:<name>
	// if namespace is not present, assume it to be cattle-global data
	split := strings.Split(o.CloudCredentialID, ":")
	name := split[0]
	namespace := "cattle-global-data"
	if len(split) > 1 {
		namespace = split[0]
		name = split[1]
	}

	cc, err := v.SecretLister.Get(namespace, name)
	if err != nil {
		return httperror.NewAPIError(httperror.NotFound, fmt.Sprintf("invalid credentials for id: %s", o.CloudCredentialID))
	}
	return o.setOCIClients(cc)
}

func (v *Validator) validateVerrazzano(errorChannel chan error, o *OCNEConfig) {
	// Don't validate Verrazzano if it's not being installed
	if !o.InstallVerrazzano {
		errorChannel <- nil
		return
	}
	if err := checkVZObject([]byte(o.VerrazzanoResource)); err != nil {
		errorChannel <- err
		return
	}

	errorChannel <- nil
}

func checkVZObject(vzBytes []byte) error {
	j, err := apiyaml.ToJSON(vzBytes)
	if err != nil {
		return httperror.NewAPIError(BadRequest, "Verrazzano resource must be valid YAML")
	}
	obj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, j)
	if err != nil {
		return httperror.NewAPIError(BadRequest, "invalid Verrazzano resource")
	}
	vz, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return httperror.NewAPIError(BadRequest, "invalid Verrazzano resource schema")
	}

	gvk := vz.GroupVersionKind()
	if gvk.Kind != "Verrazzano" {
		return httperror.NewAPIError(BadRequest, fmt.Sprintf("invalid Kind '%s' for Verrazzano", gvk.Kind))
	}
	if gvk.Group != "install.verrazzano.io" {
		return httperror.NewAPIError(BadRequest, fmt.Sprintf("invalid Group '%s' for Verrazzano", gvk.Group))
	}
	name := vz.GetName()
	namespace := vz.GetNamespace()
	if name == "" || namespace == "" {
		return httperror.NewAPIError(BadRequest, "missing required Verrazzano metadata.name and metadata.namespace")
	}
	return nil
}

func (v *Validator) validateOCNECluster(errorChannel chan error, o *OCNEConfig) {
	if err := v.validateNodePools(o); err != nil {
		errorChannel <- err
		return
	}

	errorChannel <- v.validateControlPlane(o)
}

func (v *Validator) validateNodePools(o *OCNEConfig) error {
	nodePools, err := unmarshallNodePools(o.NodePools)
	if err != nil {
		return httperror.NewAPIError(BadRequest, fmt.Sprintf("invalid node pools"))
	}

	nodePoolNames := map[string]bool{}
	for _, np := range nodePools {
		if np.Name == "" {
			return httperror.NewAPIError(BadRequest, fmt.Sprintf("missing required \"name\" field for node pool"))
		}
		if np.Replicas < 0 {
			return httperror.NewAPIError(BadRequest, fmt.Sprintf("node pool replicas must be gte 0"))
		}
		if np.Memory < 0 {
			return httperror.NewAPIError(BadRequest, fmt.Sprintf("node pool memory must be gte 0"))
		}
		if np.Ocpus < 0 {
			return httperror.NewAPIError(BadRequest, fmt.Sprintf("node pool OCPUs must be gte 0"))
		}
		if np.VolumeSize < 0 {
			return httperror.NewAPIError(BadRequest, fmt.Sprintf("node pool volume size must be gte 0"))
		}
		if nodePoolNames[np.Name] {
			return httperror.NewAPIError(BadRequest, fmt.Sprintf("duplicated node pool name \"%s\"", np.Name))
		}
		nodePoolNames[np.Name] = true
	}
	return nil
}

func (v *Validator) validateControlPlane(o *OCNEConfig) error {
	if o.NumControlPlaneNodes < 0 {
		return httperror.NewAPIError(BadRequest, fmt.Sprintf("control plane replicas must be gte 0"))
	}
	if o.ControlPlaneMemoryGbs < 0 {
		return httperror.NewAPIError(BadRequest, fmt.Sprintf("control plane memory must be gte 0"))
	}
	if o.ControlPlaneOcpus < 0 {
		return httperror.NewAPIError(BadRequest, fmt.Sprintf("control plane OCPUs must be gte 0"))
	}
	if o.ControlPlaneVolumeGbs < 0 {
		return httperror.NewAPIError(BadRequest, fmt.Sprintf("control plane size must be gte 0"))
	}
	return nil
}

func (o *OCNEConfig) setOCIClients(cc *v1.Secret) error {
	user := string(cc.Data["ocicredentialConfig-userId"])
	fingerprint := string(cc.Data["ocicredentialConfig-fingerprint"])
	tenancy := string(cc.Data["ocicredentialConfig-tenancyId"])
	passphrase := string(cc.Data["ocicredentialConfig-passphrase"])
	privateKey := string(cc.Data["ocicredentialConfig-privateKeyContents"])
	region := string(cc.Data["ocicredentialConfig-region"])
	if region == "" {
		region = o.Region
	}

	var phrase *string
	if len(passphrase) > 0 {
		phrase = &passphrase
	}
	configurationProvider := common.NewRawConfigurationProvider(tenancy, user, region, fingerprint, privateKey, phrase)
	net, err := core.NewVirtualNetworkClientWithConfigurationProvider(configurationProvider)
	if err != nil {
		return httperror.NewAPIError(httperror.Unauthorized, "failed to connect to OCI")
	}

	o.net = &net
	return nil
}

func unmarshallNodePools(serialized []string) ([]NodePool, error) {
	var nodePools []NodePool

	for _, s := range serialized {
		np := NodePool{}
		if err := json.Unmarshal([]byte(s), &np); err != nil {
			return nil, err
		}
		nodePools = append(nodePools, np)
	}

	return nodePools, nil
}

func (v *Validator) validateAdditionalYAMLs(errorChannel chan error, o *OCNEConfig) {
	for _, y := range o.ApplyYamls {
		u := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(y), u)
		if err != nil {
			errorChannel <- httperror.NewAPIError(BadRequest, "Verify Additional Cluster YAMLs are valid YAML.")
			return
		}
	}
	errorChannel <- nil
}
