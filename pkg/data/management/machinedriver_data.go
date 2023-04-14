package management

import (
	"fmt"
	errs "github.com/pkg/errors"
	v32 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/data/management/utils"
	v3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/types/config"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strings"
	"sync"
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
	driverNameLabel    = "io.cattle.node_driver.name"
)

var (
	credSchemaLock       = sync.Mutex{}
	credLock             = sync.Mutex{}
	DriverToSchemaFields = map[string]map[string]string{
		"aliyunecs":     {"sshKeypath": "sshKeyContents"},
		"amazonec2":     {"sshKeypath": "sshKeyContents", "userdata": "userdata"},
		"azure":         {"customData": "customData"},
		"digitalocean":  {"sshKeyPath": "sshKeyContents", "userdata": "userdata"},
		"exoscale":      {"sshKey": "sshKey", "userdata": "userdata"},
		"openstack":     {"cacert": "cacert", "privateKeyFile": "privateKeyFile", "userDataFile": "userDataFile"},
		"otc":           {"privateKeyFile": "privateKeyFile"},
		"packet":        {"userdata": "userdata"},
		"pod":           {"userdata": "userdata"},
		"vmwarevsphere": {"cloud-config": "cloudConfig"},
		"google":        {"authEncodedJson": "authEncodedJson"},
	}
	DriverData = map[string]map[string][]string{
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

	SSHKeyFields = map[string]bool{
		"sshKeyContents": true,
		"sshKey":         true,
		"privateKeyFile": true,
	}
)

type DynamicSchemaClients struct {
	schemaClient v3.DynamicSchemaInterface
	schemaLister v3.DynamicSchemaLister
}

func addCloudCredentials(management *config.ManagementContext) error {
	dc := DynamicSchemaClients{
		schemaClient: management.Management.DynamicSchemas(""),
		schemaLister: management.Management.DynamicSchemas("").Controller().Lister(),
	}
	if err := dc.addCloudCredential(Amazonec2driver); err != nil {
		return err
	}
	if err := dc.addCloudCredential(Azuredriver); err != nil {
		return err
	}

	if err := dc.addCloudCredential(GoogleDriver); err != nil {
		return err
	}
	return dc.addCloudCredential(OCIDriver)
}

func (d *DynamicSchemaClients) addCloudCredential(name string) error {
	// annotations can have keys cred and password, values []string to be considered as a part of cloud credential
	credLock.Lock()
	defer credLock.Unlock()

	err := errs.New("cloud credential creation failed")
	var existingSchema *v3.DynamicSchema
	annotations := map[string]string{}
	if existingSchema != nil {
		for k, v := range existingSchema.Annotations {
			annotations[k] = v
		}
	}
	for key, fields := range DriverData[name] {
		annotations[key] = strings.Join(fields, ",")
	}

	flags, err := utils.GetCreateFlagsForDriver(name)
	if err != nil {
		return err
	}
	credFields := map[string]v32.Field{}
	resourceFields := map[string]v32.Field{}

	pubCredFields, privateCredFields, passwordFields, defaults, optionals := utils.GetCredFields(annotations)
	for _, flag := range flags {

		name, field, err := utils.FlagToField(flag)
		if err != nil {
			return err
		}
		if aliases, ok := DriverToSchemaFields[name]; ok {
			// convert path fields to their alias to take file contents
			if alias, ok := aliases[name]; ok {
				name = alias
				field.Description = fmt.Sprintf("File contents for %v", alias)
			}
		}

		if privateCredFields[name] || passwordFields[name] || SSHKeyFields[name] {
			field.Type = "password"
		}

		if pubCredFields[name] || privateCredFields[name] {
			credField := field
			credField.Required = !optionals[name]
			if val, ok := defaults[name]; ok {
				credField = utils.UpdateDefault(credField, val, field.Type)
			}
			credFields[name] = credField
		}

		resourceFields[name] = field
		logrus.Infof("+++ Resource field for cloud %v = %v", name, field)
	}
	dynamicSchema := &v3.DynamicSchema{
		Spec: v32.DynamicSchemaSpec{
			ResourceFields: resourceFields,
		},
	}

	// Creating dynamic schema cloud config objects
	dynamicSchema.Name = name + "config"
	//dynamicSchema.OwnerReferences = []metav1.OwnerReference{
	//	{
	//		UID:        obj.UID,
	//		Kind:       obj.Kind,
	//		APIVersion: obj.APIVersion,
	//		Name:       obj.Name,
	//	},
	//}
	dynamicSchema.Labels = map[string]string{}
	dynamicSchema.Labels[driverNameLabel] = name

	_, err = d.schemaClient.Create(dynamicSchema)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
		ds, err := d.schemaClient.Get(dynamicSchema.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		ds.Spec.ResourceFields = resourceFields

		_, err = d.schemaClient.Update(ds)
		if err != nil {
			return err
		}
	}

	err = d.createCredSchema(name, credFields)
	if err != nil {
		return err
	}
	// Creating dynamic schema cloud credential config objects
	return d.createOrUpdateNodeForEmbeddedTypeCredential(utils.CredentialConfigSchemaName(name),
		name+"credentialConfig", true)
}

func (d *DynamicSchemaClients) createCredSchema(driverDisplayName string, credFields map[string]v32.Field) error {

	name := utils.CredentialConfigSchemaName(driverDisplayName)
	credSchema, err := d.schemaLister.Get("", name)

	if name == "amazonec2credentialconfig" {
		credFields["defaultRegion"] = v32.Field{
			Type:         "string",
			Description:  "AWS Default Region",
			DynamicField: true,
			Create:       true,
			Update:       true,
		}
	}

	if err != nil {
		if errors.IsNotFound(err) {
			credentialSchema := &v3.DynamicSchema{
				Spec: v32.DynamicSchemaSpec{
					ResourceFields: credFields,
				},
			}
			credentialSchema.Name = name
			//credentialSchema.OwnerReferences = []metav1.OwnerReference{
			//	{
			//		UID:        obj.UID,
			//		Kind:       obj.Kind,
			//		APIVersion: obj.APIVersion,
			//		Name:       obj.Name,
			//	},
			//}
			_, err := d.schemaClient.Create(credentialSchema)
			return err
		}
		return err
	} else if !reflect.DeepEqual(credSchema.Spec.ResourceFields, credFields) {
		toUpdate := credSchema.DeepCopy()
		toUpdate.Spec.ResourceFields = credFields
		_, err := d.schemaClient.Update(toUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DynamicSchemaClients) createOrUpdateNodeForEmbeddedTypeCredential(embeddedType, fieldName string, embedded bool) error {
	credSchemaLock.Lock()
	defer credSchemaLock.Unlock()

	return d.createOrUpdateNodeForEmbeddedTypeWithParents(embeddedType, fieldName, "credentialconfig", "cloudCredential", embedded, true)
}

func (d *DynamicSchemaClients) createOrUpdateNodeForEmbeddedTypeWithParents(embeddedType, fieldName, schemaID, parentID string, embedded, update bool) error {
	nodeSchema, err := d.schemaLister.Get("", schemaID)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if errors.IsNotFound(err) {
		resourceField := map[string]v32.Field{}
		if embedded {
			resourceField[fieldName] = v32.Field{
				Create:   true,
				Nullable: true,
				Update:   update,
				Type:     embeddedType,
			}
		}
		dynamicSchema := &v3.DynamicSchema{}
		dynamicSchema.Name = schemaID
		dynamicSchema.Spec.ResourceFields = resourceField
		dynamicSchema.Spec.Embed = true
		dynamicSchema.Spec.EmbedType = parentID
		_, err := d.schemaClient.Create(dynamicSchema)
		if err != nil {
			return err
		}
		return nil
	}

	nodeSchema = nodeSchema.DeepCopy()

	shouldUpdate := false
	if embedded {
		if nodeSchema.Spec.ResourceFields == nil {
			nodeSchema.Spec.ResourceFields = map[string]v32.Field{}
		}
		if _, ok := nodeSchema.Spec.ResourceFields[fieldName]; !ok {
			// if embedded we add the type to schema
			logrus.Infof("uploading %s to %s schema", fieldName, schemaID)
			nodeSchema.Spec.ResourceFields[fieldName] = v32.Field{
				Create:   true,
				Nullable: true,
				Update:   update,
				Type:     embeddedType,
			}
			shouldUpdate = true
		}
	} else {
		// if not we delete it from schema
		if _, ok := nodeSchema.Spec.ResourceFields[fieldName]; ok {
			logrus.Infof("deleting %s from %s schema", fieldName, schemaID)
			delete(nodeSchema.Spec.ResourceFields, fieldName)
			shouldUpdate = true
		}
	}

	if shouldUpdate {
		_, err = d.schemaClient.Update(nodeSchema)
		if err != nil {
			return err
		}
	}

	return nil
}
