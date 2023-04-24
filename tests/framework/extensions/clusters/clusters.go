// Copyright (c) 2023, Oracle and/or its affiliates.

// This file from the Rancher repository has been modified by Oracle as follows:
// - references to the etcdsnapshots.rke.cattle.io CRDs and APIs have been removed

package clusters

import (
	"fmt"
	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/pkg/api/scheme"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/rancher/tests/framework/clients/rancher"
	management "github.com/rancher/rancher/tests/framework/clients/rancher/generated/management/v3"
	"github.com/rancher/rancher/tests/framework/pkg/wait"
	"github.com/rancher/rancher/tests/integration/pkg/defaults"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

// GetClusterIDByName is a helper function that returns the cluster ID by name
func GetClusterIDByName(client *rancher.Client, clusterName string) (string, error) {
	clusterList, err := client.Management.Cluster.List(&types.ListOpts{})
	if err != nil {
		return "", err
	}

	for _, cluster := range clusterList.Data {
		if cluster.Name == clusterName {
			return cluster.ID, nil
		}
	}

	return "", nil
}

// IsHostedProvisioningClusterReady is basic check function that would be used for the wait.WatchWait func in pkg/wait.
// This functions just waits until a hosted cluster becomes ready.
func IsHostedProvisioningClusterReady(event watch.Event) (ready bool, err error) {
	clusterUnstructured := event.Object.(*unstructured.Unstructured)
	cluster := &v3.Cluster{}
	err = scheme.Scheme.Convert(clusterUnstructured, cluster, clusterUnstructured.GroupVersionKind())
	if err != nil {
		return false, err
	}
	for _, cond := range cluster.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true, nil
		}
	}

	return false, nil
}

// Verify if a serviceAccountTokenSecret exists or not in the cluster.
func CheckServiceAccountTokenSecret(client *rancher.Client, clusterName string) (success bool, err error) {
	clusterID, err := GetClusterIDByName(client, clusterName)
	if err != nil {
		return false, err
	}

	cluster, err := client.Management.Cluster.ByID(clusterID)
	if err != nil {
		return false, err
	}

	if cluster.ServiceAccountTokenSecret != "" {
		logrus.Infof("serviceAccountTokenSecret in this cluster is: %s", cluster.ServiceAccountTokenSecret)
		return true, nil
	} else {
		logrus.Infof("serviceAccountTokenSecret does not exist in this cluster!")
		return false, nil
	}
}

// NewRKE1lusterConfig is a constructor for a v3.Cluster object, to be used by the rancher.Client.Provisioning client.
func NewRKE1ClusterConfig(clusterName, cni, kubernetesVersion string, client *rancher.Client) *management.Cluster {
	clusterConfig := &management.Cluster{
		DockerRootDir:           "/var/lib/docker",
		EnableClusterMonitoring: false,
		LocalClusterAuthEndpoint: &management.LocalClusterAuthEndpoint{
			Enabled: true,
		},
		Name: clusterName,
		RancherKubernetesEngineConfig: &management.RancherKubernetesEngineConfig{
			DNS: &management.DNSConfig{
				Provider: "coredns",
				Options: map[string]string{
					"stubDomains": "cluster.local",
				},
			},
			Ingress: &management.IngressConfig{
				Provider: "nginx",
			},
			Monitoring: &management.MonitoringConfig{
				Provider: "metrics-server",
			},
			Network: &management.NetworkConfig{
				MTU:     0,
				Options: map[string]string{},
			},
			Version: kubernetesVersion,
		},
	}

	return clusterConfig
}

// CreateRKE1Cluster is a "helper" functions that takes a rancher client, and the rke1 cluster config as parameters. This function
// registers a delete cluster fuction with a wait.WatchWait to ensure the cluster is removed cleanly.
func CreateRKE1Cluster(client *rancher.Client, rke1Cluster *management.Cluster) (*management.Cluster, error) {
	opts := metav1.ListOptions{
		FieldSelector:  "metadata.name=",
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	}

	cluster, err := client.Management.Cluster.Create(rke1Cluster)
	if err != nil {
		return nil, err
	}

	client, err = client.ReLogin()
	if err != nil {
		return nil, err
	}

	client.Session.RegisterCleanupFunc(func() error {
		err := client.Management.Cluster.Delete(rke1Cluster)
		if err != nil {
			return err
		}

		adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
		if err != nil {
			return err
		}

		watchInterface, err := adminClient.GetManagementWatchInterface(management.ClusterType, opts)

		if err != nil {
			return err
		}

		return wait.WatchWait(watchInterface, func(event watch.Event) (ready bool, err error) {
			if event.Type == watch.Error {
				return false, fmt.Errorf("there was an error deleting cluster")
			} else if event.Type == watch.Deleted {
				return true, nil
			}
			return false, nil
		})
	})

	return cluster, nil
}
