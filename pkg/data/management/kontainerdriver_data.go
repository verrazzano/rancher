package management

import (
	"context"
	"fmt"
	"os"
	"strings"

	v32 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/controllers/management/drivers/kontainerdriver"
	v3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/namespace"
	"github.com/rancher/rancher/pkg/types/config"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func addKontainerDrivers(management *config.ManagementContext) error {
	// create binary drop location if not exists
	err := os.MkdirAll(kontainerdriver.DriverDir, 0777)
	if err != nil {
		return fmt.Errorf("error creating binary drop folder: %v", err)
	}

	creator := driverCreator{
		driversLister: management.Management.KontainerDrivers("").Controller().Lister(),
		drivers:       management.Management.KontainerDrivers(""),
		k8s:           management.K8sClient,
	}

	if err := cleanupImportDriver(creator); err != nil {
		return err
	}

	if err := removeUnsupportedDrivers(creator); err != nil {
		return err
	}

	if err := creator.add("googleKubernetesEngine"); err != nil {
		return err
	}

	if err := creator.add("azureKubernetesService"); err != nil {
		return err
	}

	if err := creator.add("amazonElasticContainerService"); err != nil {
		return err
	}

	if err := creator.addHostedDriverFromEnv("ociocne", "OCI_OCNE_DRIVER_VERSION", "OCI_OCNE_DRIVER_HASH"); err != nil {
		return err
	}

	return creator.addCustomDriver(
		"oraclecontainerengine",
		"https://github.com/rancher-plugins/kontainer-engine-driver-oke/releases/download/v1.8.3/kontainer-engine-driver-oke-linux",
		"7bfde567e6d478f1da8d36531f765d348bff1cd3abe83c70ddf7766f46112170",
		"",
		false,
		"*.oraclecloud.com",
	)

}

func cleanupImportDriver(creator driverCreator) error {
	var err error
	if _, err = creator.driversLister.Get("", "import"); err == nil {
		err = creator.drivers.Delete("import", &v1.DeleteOptions{})
	}

	if !errors.IsNotFound(err) {
		return err
	}

	return nil
}

// removeUnsupportedDrivers - remove unsupported drivers
func removeUnsupportedDrivers(creator driverCreator) error {
	driverList := []string{"aliyunkubernetescontainerservice", "baiducloudcontainerengine", "huaweicontainercloudengine",
		"linodekubernetesengine", "opentelekomcloudcontainerengine", "rancherkubernetesengine", "tencentkubernetesengine"}

	var err error

	for _, driver := range driverList {
		if _, err = creator.driversLister.Get("", driver); err == nil {
			logrus.Infof("removing kontainer drvier %s", driver)
			err = creator.drivers.Delete(driver, &v1.DeleteOptions{})
		}
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

type driverCreator struct {
	driversLister v3.KontainerDriverLister
	drivers       v3.KontainerDriverInterface
	k8s           kubernetes.Interface
}

func (c *driverCreator) add(name string) error {
	logrus.Infof("adding kontainer driver %s", name)

	driver, err := c.driversLister.Get("", name)
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = c.drivers.Create(&v3.KontainerDriver{
				ObjectMeta: v1.ObjectMeta{
					Name:      strings.ToLower(name),
					Namespace: "",
				},
				Spec: v32.KontainerDriverSpec{
					URL:     "",
					BuiltIn: true,
					Active:  true,
				},
				Status: v32.KontainerDriverStatus{
					DisplayName: name,
				},
			})
			if err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("error creating driver: %v", err)
			}
		} else {
			return fmt.Errorf("error getting driver: %v", err)
		}
	} else {
		driver.Spec.URL = ""

		_, err = c.drivers.Update(driver)
		if err != nil {
			return fmt.Errorf("error updating driver: %v", err)
		}
	}

	return nil
}

func (c *driverCreator) addHostedDriverFromEnv(name, versionEnv, checksumEnv string, domains ...string) error {
	version := os.Getenv(versionEnv)
	checksum := os.Getenv(checksumEnv)
	// don't add driver if not present in environment
	if version == "" || checksum == "" {
		return nil
	}

	ingress, err := c.k8s.NetworkingV1().Ingresses(namespace.System).Get(context.Background(), "rancher", v1.GetOptions{})
	if err != nil {
		return err
	}

	if ingress.Annotations != nil {
		if commonName, ok := ingress.Annotations["cert-manager.io/common-name"]; ok {
			url := fmt.Sprintf("https://%s/kontainerdriver/%s/%s/kontainer-engine-driver-%s-linux", commonName, name, version, name)
			return c.addCustomDriver(fmt.Sprintf("%sengine", name), url, checksum, "", true, domains...)
		}
	}

	return fmt.Errorf("failed to create hosted driver, %s/rancher ingress not ready", namespace.System)
}

func (c *driverCreator) addCustomDriver(name, url, checksum, uiURL string, active bool, domains ...string) error {
	logrus.Infof("adding kontainer driver %v", name)
	driver, err := c.driversLister.Get("", name)
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = c.drivers.Create(&v3.KontainerDriver{
				ObjectMeta: v1.ObjectMeta{
					Name: strings.ToLower(name),
				},
				Spec: v32.KontainerDriverSpec{
					URL:              url,
					BuiltIn:          false,
					Active:           active,
					Checksum:         checksum,
					UIURL:            uiURL,
					WhitelistDomains: domains,
				},
				Status: v32.KontainerDriverStatus{
					DisplayName: name,
				},
			})
			if err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("error creating driver: %v", err)
			}
			return nil
		}

		return fmt.Errorf("error getting driver: %v", err)
	}

	// do an update if the driver already exists
	driver.Spec.URL = url
	driver.Spec.BuiltIn = false
	driver.Spec.Active = active
	driver.Spec.Checksum = checksum
	driver.Spec.UIURL = uiURL
	driver.Spec.WhitelistDomains = domains

	_, err = c.drivers.Update(driver)
	if err != nil {
		return fmt.Errorf("error updating driver: %v", err)
	}
	return nil
}
