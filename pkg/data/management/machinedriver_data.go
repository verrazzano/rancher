package management

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
