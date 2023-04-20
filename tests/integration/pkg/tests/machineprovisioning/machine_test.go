package machineprovisioning

import (
	"os"
	"strings"
	"testing"

	provisioningv1api "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2"
	"github.com/rancher/rancher/tests/integration/pkg/clients"
	"github.com/rancher/rancher/tests/integration/pkg/cluster"
	"github.com/rancher/rancher/tests/integration/pkg/defaults"
	"github.com/rancher/rancher/tests/integration/pkg/nodeconfig"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

func TestSingleNodeAllRolesWithDelete(t *testing.T) {
	clients, err := clients.New()
	if err != nil {
		t.Fatal(err)
	}
	defer clients.Close()

	c, err := cluster.New(clients, &provisioningv1api.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-single-node-all-roles-with-delete",
		},
		Spec: provisioningv1api.ClusterSpec{
			KubernetesVersion: defaults.SomeK8sVersion,
			RKEConfig: &provisioningv1api.RKEConfig{
				MachinePools: []provisioningv1api.RKEMachinePool{{
					EtcdRole:         true,
					ControlPlaneRole: true,
					WorkerRole:       true,
					Quantity:         &defaults.One,
				}},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForCreate(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	machines, err := cluster.Machines(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(machines.Items), 1)
	machine := machines.Items[0]

	clusterClients, err := clients.ForCluster(c.Namespace, c.Name)
	if err != nil {
		t.Fatal(err)
	}

	node, err := clusterClients.Core.Node().Get(machines.Items[0].Status.NodeRef.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	args, err := nodeconfig.FromNode(node)
	if err != nil {
		t.Fatal(err)
	}

	// This shouldn't be one, fix when node args starts returning what is from the config file
	assert.Greater(t, len(args), 10)
	assert.Len(t, machine.Status.Addresses, 2)

	// Delete the cluster and wait for cleanup.
	err = clients.Provisioning.Cluster().Delete(c.Namespace, c.Name, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForDelete(clients, c)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMachineTemplateClonedAnnotations(t *testing.T) {
	if strings.ToLower(os.Getenv("DIST")) == "rke2" {
		t.Skip()
	}

	clients, err := clients.New()
	if err != nil {
		t.Fatal(err)
	}
	defer clients.Close()

	c, err := cluster.New(clients, &provisioningv1api.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine-template-cloned-annotations",
		},
		Spec: provisioningv1api.ClusterSpec{
			KubernetesVersion: defaults.SomeK8sVersion,
			RKEConfig: &provisioningv1api.RKEConfig{
				MachinePools: []provisioningv1api.RKEMachinePool{{
					EtcdRole:         true,
					ControlPlaneRole: true,
					WorkerRole:       true,
					Quantity:         &defaults.One,
				}},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForCreate(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	infraMachines, err := cluster.PodInfraMachines(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	for _, infraMachine := range infraMachines.Items {
		templateGroupKind := schema.ParseGroupKind(infraMachine.GetAnnotations()[capi.TemplateClonedFromGroupKindAnnotation])

		machineTemplate, err := clients.Dynamic.
			Resource(schema.GroupVersionResource{Group: templateGroupKind.Group, Version: "v1", Resource: strings.ToLower(templateGroupKind.Kind) + "s"}).
			Namespace(infraMachine.GetNamespace()).
			Get(clients.Ctx, infraMachine.GetAnnotations()[capi.TemplateClonedFromNameAnnotation], metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		gv, err := schema.ParseGroupVersion(machineTemplate.GetAnnotations()[rke2.MachineTemplateClonedFromGroupVersionAnn])
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, gv.String(), rke2.DefaultMachineConfigAPIVersion)
		assert.Equal(t, machineTemplate.GetAnnotations()[rke2.MachineTemplateClonedFromKindAnn], c.Spec.RKEConfig.MachinePools[0].NodeConfig.Kind)
		assert.Equal(t, machineTemplate.GetAnnotations()[rke2.MachineTemplateClonedFromNameAnn], c.Spec.RKEConfig.MachinePools[0].NodeConfig.Name)
	}
}

func TestMachineSetDeletePolicyOldestSet(t *testing.T) {
	if strings.ToLower(os.Getenv("DIST")) == "rke2" {
		t.Skip()
	}

	clients, err := clients.New()
	if err != nil {
		t.Fatal(err)
	}
	defer clients.Close()

	c, err := cluster.New(clients, &provisioningv1api.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine-set-delete-policy-oldest-set",
		},
		Spec: provisioningv1api.ClusterSpec{
			KubernetesVersion: defaults.SomeK8sVersion,
			RKEConfig: &provisioningv1api.RKEConfig{
				MachinePools: []provisioningv1api.RKEMachinePool{
					{
						EtcdRole:         true,
						ControlPlaneRole: true,
						WorkerRole:       true,
						Quantity:         &defaults.One,
					},
					{
						EtcdRole:         true,
						ControlPlaneRole: true,
						WorkerRole:       true,
						Quantity:         &defaults.One,
						RollingUpdate: &provisioningv1api.RKEMachinePoolRollingUpdate{
							MaxUnavailable: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "10%",
							},
							MaxSurge: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "10%",
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForCreate(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	machineSets, err := cluster.MachineSets(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	for _, machineSet := range machineSets.Items {
		d, err := data.Convert(machineSet)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, string(capi.OldestMachineSetDeletePolicy), d.String("Object", "spec", "deletePolicy"))
	}
}

func TestThreeNodesAllRolesWithDelete(t *testing.T) {
	clients, err := clients.New()
	if err != nil {
		t.Fatal(err)
	}
	defer clients.Close()

	c, err := cluster.New(clients, &provisioningv1api.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-three-nodes-all-roles-with-delete",
		},
		Spec: provisioningv1api.ClusterSpec{
			KubernetesVersion: defaults.SomeK8sVersion,
			RKEConfig: &provisioningv1api.RKEConfig{
				MachinePools: []provisioningv1api.RKEMachinePool{{
					EtcdRole:         true,
					ControlPlaneRole: true,
					WorkerRole:       true,
					Quantity:         &defaults.Three,
				}},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForCreate(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	// Delete the cluster and wait for cleanup.
	err = clients.Provisioning.Cluster().Delete(c.Namespace, c.Name, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForDelete(clients, c)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFiveNodesUniqueRolesWithDelete(t *testing.T) {
	if strings.ToLower(os.Getenv("DIST")) == "rke2" {
		t.Skip()
	}
	clients, err := clients.New()
	if err != nil {
		t.Fatal(err)
	}
	defer clients.Close()

	c, err := cluster.New(clients, &provisioningv1api.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-five-nodes-unique-roles-with-delete",
		},
		Spec: provisioningv1api.ClusterSpec{
			KubernetesVersion: defaults.SomeK8sVersion,
			RKEConfig: &provisioningv1api.RKEConfig{
				MachinePools: []provisioningv1api.RKEMachinePool{
					{
						EtcdRole: true,
						Quantity: &defaults.Three,
					},
					{
						ControlPlaneRole: true,
						Quantity:         &defaults.One,
					},
					{
						WorkerRole: true,
						Quantity:   &defaults.One,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForCreate(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	// Delete the cluster and wait for cleanup.
	err = clients.Provisioning.Cluster().Delete(c.Namespace, c.Name, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForDelete(clients, c)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFourNodesServerAndWorkerRolesWithDelete(t *testing.T) {
	if strings.ToLower(os.Getenv("DIST")) == "rke2" {
		t.Skip()
	}
	t.Parallel()
	clients, err := clients.New()
	if err != nil {
		t.Fatal(err)
	}
	defer clients.Close()

	c, err := cluster.New(clients, &provisioningv1api.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-four-nodes-server-and-worker-roles-with-delete",
		},
		Spec: provisioningv1api.ClusterSpec{
			KubernetesVersion: defaults.SomeK8sVersion,
			RKEConfig: &provisioningv1api.RKEConfig{
				MachinePools: []provisioningv1api.RKEMachinePool{
					{
						EtcdRole:         true,
						ControlPlaneRole: true,
						Quantity:         &defaults.Three,
					},
					{
						WorkerRole: true,
						Quantity:   &defaults.One,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForCreate(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	// Delete the cluster and wait for cleanup.
	err = clients.Provisioning.Cluster().Delete(c.Namespace, c.Name, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForDelete(clients, c)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDrainNoDelete(t *testing.T) {
	clients, err := clients.New()
	if err != nil {
		t.Fatal(err)
	}
	defer clients.Close()

	c, err := cluster.New(clients, &provisioningv1api.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-drain-no-delete",
		},
		Spec: provisioningv1api.ClusterSpec{
			KubernetesVersion: defaults.SomeK8sVersion,
			RKEConfig: &provisioningv1api.RKEConfig{
				MachinePools: []provisioningv1api.RKEMachinePool{
					{
						EtcdRole:          true,
						ControlPlaneRole:  true,
						Quantity:          &defaults.One,
						DrainBeforeDelete: false,
					},
					{
						WorkerRole:        true,
						Quantity:          &defaults.One,
						DrainBeforeDelete: true,
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	c, err = cluster.WaitForCreate(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	machines, err := cluster.Machines(clients, c)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(machines.Items), 2)

	excludeNodeDraining, ok := machines.Items[0].Annotations[capi.ExcludeNodeDrainingAnnotation]
	assert.True(t, ok)
	assert.Equal(t, excludeNodeDraining, "true")

	_, ok = machines.Items[1].Annotations[capi.ExcludeNodeDrainingAnnotation]
	assert.False(t, ok)
}
