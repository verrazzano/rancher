package management

import (
	"fmt"
	v32 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	v3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/types/config"
	v1 "github.com/rancher/rancher/tests/framework/clients/rancher/v1"
	"github.com/sirupsen/logrus"
	"os/exec"
	"reflect"
	"strings"
)

const (
	Amazonec2driver    = "amazonec2"
	Azuredriver        = "azure"
	DigitalOceandriver = "digitalocean"
	ExoscaleDriver     = "exoscale"
	HarvesterDriver    = "harvester"
	Linodedriver       = "linode"
	NutanixDriver      = "nutanix"
	OCIDriver          = "oci"
	OTCDriver          = "otc"
	OpenstackDriver    = "openstack"
	PacketDriver       = "packet"
	PhoenixNAPDriver   = "pnap"
	RackspaceDriver    = "rackspace"
	SoftLayerDriver    = "softlayer"
	Vmwaredriver       = "vmwarevsphere"
	GoogleDriver       = "google"
	OutscaleDriver     = "outscale"
)

var DriverData = map[string]map[string][]string{
	Amazonec2driver:    {"publicCredentialFields": []string{"accessKey"}, "privateCredentialFields": []string{"secretKey"}},
	Azuredriver:        {"publicCredentialFields": []string{"clientId", "subscriptionId", "tenantId", "environment"}, "privateCredentialFields": []string{"clientSecret"}, "optionalCredentialFields": []string{"tenantId"}},
	DigitalOceandriver: {"privateCredentialFields": []string{"accessToken"}},
	ExoscaleDriver:     {"privateCredentialFields": []string{"apiSecretKey"}},
	HarvesterDriver:    {"publicCredentialFields": []string{"clusterType", "clusterId"}, "privateCredentialFields": []string{"kubeconfigContent"}, "optionalCredentialFields": []string{"clusterId"}},
	Linodedriver:       {"privateCredentialFields": []string{"token"}, "passwordFields": []string{"rootPass"}},
	NutanixDriver:      {"publicCredentialFields": []string{"endpoint", "username", "port"}, "privateCredentialFields": []string{"password"}},
	OCIDriver:          {"publicCredentialFields": []string{"tenancyId", "userId", "fingerprint"}, "privateCredentialFields": []string{"privateKeyContents"}, "passwordFields": []string{"privateKeyPassphrase"}},
	OTCDriver:          {"privateCredentialFields": []string{"accessKeySecret"}},
	OpenstackDriver:    {"privateCredentialFields": []string{"password"}},
	PacketDriver:       {"privateCredentialFields": []string{"apiKey"}},
	PhoenixNAPDriver:   {"publicCredentialFields": []string{"clientIdentifier"}, "privateCredentialFields": []string{"clientSecret"}},
	RackspaceDriver:    {"privateCredentialFields": []string{"apiKey"}},
	SoftLayerDriver:    {"privateCredentialFields": []string{"apiKey"}},
	Vmwaredriver:       {"publicCredentialFields": []string{"username", "vcenter", "vcenterPort"}, "privateCredentialFields": []string{"password"}},
	GoogleDriver:       {"privateCredentialFields": []string{"authEncodedJson"}},
	OutscaleDriver:     {"publicCredentialFields": []string{"accessKey", "region"}, "privateCredentialFields": []string{"secretKey"}},
}

var driverDefaults = map[string]map[string]string{
	HarvesterDriver: {"clusterType": "imported"},
	Vmwaredriver:    {"vcenterPort": "443"},
}

type machineDriverCompare struct {
	builtin            bool
	addCloudCredential bool
	url                string
	uiURL              string
	checksum           string
	name               string
	whitelist          []string
	annotations        map[string]string
}

func addMachineDrivers(management *config.ManagementContext) error {
	if err := addMachineDriver(Amazonec2driver, "local://", "", "",
		[]string{"iam.amazonaws.com", "iam.us-gov.amazonaws.com", "iam.%.amazonaws.com.cn", "ec2.%.amazonaws.com", "ec2.%.amazonaws.com.cn", "eks.%.amazonaws.com", "eks.%.amazonaws.com.cn", "kms.%.amazonaws.com", "kms.%.amazonaws.com.cn"},
		true, true, false, management); err != nil {
		return err
	}
	if err := addMachineDriver(Azuredriver, "local://", "", "", nil, true, true, false, management); err != nil {
		return err
	}

	return addMachineDriver(GoogleDriver, "local://", "", "", nil, false, true, true, management)
}

func addMachineDriver(name, url, uiURL, checksum string, whitelist []string, active, builtin, addCloudCredential bool, management *config.ManagementContext) error {
	lister := management.Management.NodeDrivers("").Controller().Lister()
	cli := management.Management.NodeDrivers("")
	m, _ := lister.Get("", name)
	// annotations can have keys cred and password, values []string to be considered as a part of cloud credential
	annotations := map[string]string{}
	if m != nil {
		for k, v := range m.Annotations {
			annotations[k] = v
		}
	}
	for key, fields := range DriverData[name] {
		annotations[key] = strings.Join(fields, ",")
	}
	defaults := []string{}
	for key, val := range driverDefaults[name] {
		defaults = append(defaults, fmt.Sprintf("%s:%s", key, val))
	}
	if len(defaults) > 0 {
		annotations["defaults"] = strings.Join(defaults, ",")
	}
	if m != nil {
		old := machineDriverCompare{
			builtin:            m.Spec.Builtin,
			addCloudCredential: m.Spec.AddCloudCredential,
			url:                m.Spec.URL,
			uiURL:              m.Spec.UIURL,
			checksum:           m.Spec.Checksum,
			name:               m.Spec.DisplayName,
			whitelist:          m.Spec.WhitelistDomains,
			annotations:        m.Annotations,
		}
		new := machineDriverCompare{
			builtin:            builtin,
			addCloudCredential: addCloudCredential,
			url:                url,
			uiURL:              uiURL,
			checksum:           checksum,
			name:               name,
			whitelist:          whitelist,
			annotations:        annotations,
		}
		if !reflect.DeepEqual(new, old) {
			logrus.Infof("Updating node driver %v", name)
			m.Spec.Builtin = builtin
			m.Spec.AddCloudCredential = addCloudCredential
			m.Spec.URL = url
			m.Spec.UIURL = uiURL
			m.Spec.Checksum = checksum
			m.Spec.DisplayName = name
			m.Spec.WhitelistDomains = whitelist
			m.Annotations = annotations
			_, err := cli.Update(m)
			return err
		}
		return nil
	}

	logrus.Infof("Creating node driver %v", name)
	_, err := cli.Create(&v3.NodeDriver{
		ObjectMeta: v1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
		},
		Spec: v32.NodeDriverSpec{
			Active:             active,
			Builtin:            builtin,
			AddCloudCredential: addCloudCredential,
			URL:                url,
			UIURL:              uiURL,
			DisplayName:        name,
			Checksum:           checksum,
			WhitelistDomains:   whitelist,
		},
	})

	return err
}

func isCommandAvailable(name string) bool {
	return exec.Command("command", "-v", name).Run() == nil
}
