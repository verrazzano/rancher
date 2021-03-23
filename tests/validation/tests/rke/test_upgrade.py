import os

from .conftest import *  # NOQA
from .common import *  # NOQA
import pytest

K8S_PREUPGRADE_IMAGE = os.environ.get(
    'RANCHER_K8S_PREUPGRADE_IMAGE', 'v1.16.8-rancher1-3')
K8S_UPGRADE_IMAGE = os.environ.get(
    'RANCHER_K8S_UPGRADE_IMAGE', 'v1.17.4-rancher1-3')


def test_upgrade_1(test_name, cloud_provider, rke_client, kubectl):
    """
    Update cluster k8s service images, cluster config:
    node0 - controlplane, etcd
    node1 - worker
    node2 - worker
    """
    print(K8S_UPGRADE_IMAGE)
    print(K8S_PREUPGRADE_IMAGE)

    rke_template = 'cluster_upgrade_1_1.yml.j2'
    all_nodes = cloud_provider.create_multiple_nodes(3, test_name)

    rke_config = create_rke_cluster(
        rke_client, kubectl, all_nodes, rke_template,
        k8_rancher_image=K8S_PREUPGRADE_IMAGE)

    network, dns_discovery = validate_rke_cluster(
        rke_client, kubectl, all_nodes, 'beforeupgrade')

    validate_k8s_service_images(all_nodes, K8S_PREUPGRADE_IMAGE,
                                rke_client, kubectl)

    # New cluster needs to keep controlplane and etcd nodes the same
    rke_config = create_rke_cluster(
        rke_client, kubectl, all_nodes, rke_template,
        k8_rancher_image=K8S_UPGRADE_IMAGE)
    # The updated images versions need to be validated
    validate_k8s_service_images(all_nodes, K8S_UPGRADE_IMAGE,
                                rke_client, kubectl)
    # Rerun validation on existing and new resources
    validate_rke_cluster(
        rke_client, kubectl, all_nodes, 'beforeupgrade',
        network_validation=network, dns_validation=dns_discovery)
    validate_rke_cluster(
        rke_client, kubectl, all_nodes, 'afterupgrade')
    delete_nodes(cloud_provider, all_nodes)


def test_upgrade_2(test_name, cloud_provider, rke_client, kubectl):
    """
    Update cluster k8s service images and add worker node, cluster config:
    node0 - controlplane, etcd
    node1 - worker
    node2 - worker
    Upgrade adds a worker node:
    node0 - controlplane, etcd
    node1 - worker
    node2 - worker
    node3 - worker
    """
    rke_template = 'cluster_upgrade_2_1.yml.j2'
    all_nodes = cloud_provider.create_multiple_nodes(4, test_name)
    before_upgrade_nodes = all_nodes[0:-1]
    rke_config = create_rke_cluster(
        rke_client, kubectl, before_upgrade_nodes, rke_template,
        k8_rancher_image=K8S_PREUPGRADE_IMAGE)
    network, dns_discovery = validate_rke_cluster(
        rke_client, kubectl, before_upgrade_nodes, 'beforeupgrade')
    validate_k8s_service_images(before_upgrade_nodes, K8S_PREUPGRADE_IMAGE,
                                rke_client, kubectl)

    # New cluster needs to keep controlplane and etcd nodes the same
    rke_template = 'cluster_upgrade_2_2.yml.j2'
    rke_config = create_rke_cluster(
        rke_client, kubectl, all_nodes, rke_template,
        k8_rancher_image=K8S_UPGRADE_IMAGE)
    validate_k8s_service_images(all_nodes, K8S_UPGRADE_IMAGE,
                                rke_client, kubectl)

    # Rerun validation on existing and new resources
    validate_rke_cluster(
        rke_client, kubectl, all_nodes, 'beforeupgrade',
        network_validation=network, dns_validation=dns_discovery)
    validate_rke_cluster(
        rke_client, kubectl, all_nodes, 'afterupgrade')
    delete_nodes(cloud_provider, all_nodes)


def test_upgrade_3(test_name, cloud_provider, rke_client, kubectl):
    """
    Update cluster k8s service images and remove worker node, cluster config:
    node0 - controlplane, etcd
    node1 - worker
    node2 - worker
    Upgrade removes a worker node:
    node0 - controlplane, etcd
    node1 - worker
    """
    rke_template = 'cluster_upgrade_3_1.yml.j2'
    all_nodes = cloud_provider.create_multiple_nodes(3, test_name)
    rke_config = create_rke_cluster(
        rke_client, kubectl, all_nodes, rke_template,
        k8_rancher_image=K8S_PREUPGRADE_IMAGE)
    network, dns_discovery = validate_rke_cluster(
        rke_client, kubectl, all_nodes, 'beforeupgrade')
    validate_k8s_service_images(all_nodes, K8S_PREUPGRADE_IMAGE,
                                rke_client, kubectl)

    # New cluster needs to keep controlplane and etcd nodes the same
    rke_template = 'cluster_upgrade_3_2.yml.j2'
    after_upgrade_nodes = all_nodes[0:-1]
    rke_config = create_rke_cluster(
        rke_client, kubectl, after_upgrade_nodes, rke_template,
        k8_rancher_image=K8S_UPGRADE_IMAGE)
    validate_k8s_service_images(after_upgrade_nodes, K8S_UPGRADE_IMAGE,
                                rke_client, kubectl)

    # Rerun validation on existing and new resources
    validate_rke_cluster(
        rke_client, kubectl, after_upgrade_nodes, 'beforeupgrade',
        network_validation=network, dns_validation=dns_discovery)
    validate_rke_cluster(
        rke_client, kubectl, after_upgrade_nodes, 'afterupgrade')
    delete_nodes(cloud_provider, all_nodes)


@pytest.mark.skipif(
    True, reason="This test is skipped for now")
def test_upgrade_4(test_name, cloud_provider, rke_client, kubectl):
    """
    Update only one service in cluster k8s system images, cluster config:
    node0 - controlplane, etcd
    node1 - worker
    node2 - worker
    """
    rke_template = 'cluster_upgrade_4_1.yml.j2'
    all_nodes = cloud_provider.create_multiple_nodes(3, test_name)
    rke_config = create_rke_cluster(
        rke_client, kubectl, all_nodes, rke_template,
        k8_rancher_image=K8S_PREUPGRADE_IMAGE)
    network, dns_discovery = validate_rke_cluster(
        rke_client, kubectl, all_nodes, 'beforeupgrade')
    validate_k8s_service_images(all_nodes, K8S_PREUPGRADE_IMAGE,
                                rke_client, kubectl)

    # Upgrade only the scheduler, yaml will replace `upgrade_k8_rancher_image`
    # for scheduler image only, the rest will keep pre-upgrade image
    rke_config = create_rke_cluster(
        rke_client, kubectl, all_nodes, rke_template,
        k8_rancher_image=K8S_PREUPGRADE_IMAGE,
        upgrade_k8_rancher_image=K8S_UPGRADE_IMAGE)
    validate_k8s_service_images(all_nodes, K8S_UPGRADE_IMAGE,
                                rke_client, kubectl)
    # Rerun validation on existing and new resources
    validate_rke_cluster(
        rke_client, kubectl, all_nodes, 'beforeupgrade',
        network_validation=network, dns_validation=dns_discovery)
    validate_rke_cluster(
        rke_client, kubectl, all_nodes, 'afterupgrade')

    delete_nodes(cloud_provider, all_nodes)
