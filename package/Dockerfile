# Copyright (C) 2023, Oracle and/or its affiliates.

# This file from the Rancher repository has been modified by Oracle as follows:
# - references to k3s have been removed

FROM ghcr.io/oracle/oraclelinux:7-slim

ARG ARCH=amd64

RUN yum-config-manager --enable ol7_optional_latest && \
    yum-config-manager --enable ol7_addons && \
    yum-config-manager --add-repo=http://yum.oracle.com/repo/OracleLinux/OL7/SoftwareCollections/x86_64 && \
    yum-config-manager --add-repo https://yum.oracle.com/repo/OracleLinux/OL7/olcne13/x86_64/ && \
    yum-config-manager --add-repo https://yum.oracle.com/repo/OracleLinux/OL7/olcne15/x86_64/ && \
    yum-config-manager --add-repo https://yum.oracle.com/repo/OracleLinux/OL7/developer_EPEL/x86_64 && \
    yum-config-manager --enable ol7_developer_golang119 && \
    yum install oracle-golang-release-el7 -y && \
    yum clean all

RUN yum update -y && \
    # software collections repo needed for git 2.x on OL7
    yum -y install rh-git227 curl ca-certificates unzip tar nc mkisofs helm-3.3.4-2.el7.x86_64 golang-1.19 docker-engine-19.03.11.ol-13.el7.x86_64 kubectl-1.23.11-1.el7.x86_64 etcd-3.5.1-2.el7.x86_64 tini-0.19.0-2.el7.x86_64 && \
    # Set up needed to ensure git 2.27 from rh-git227 is on the path
    ln /opt/rh/rh-git227/enable /etc/profile.d/git.sh && \
    source /etc/profile.d/git.sh && \
    git version && \
    useradd rancher && \
    rm -rf /tmp/* /var/tmp/* /usr/share/doc/packages/* && \
    mkdir -p /var/lib/rancher /var/lib/cattle /opt/jail /opt/drivers/management-state/bin && \
    chown -R rancher /var/lib/rancher /var/lib/cattle /usr/local/bin && \
    ln -s /usr/bin/rancher /usr/bin/reset-password && \
    ln -s /usr/bin/rancher /usr/bin/ensure-default-admin && \
    rm -f /bin/sh && ln -s /bin/bash /bin/sh && \
    mv /usr/bin/helm /usr/bin/helm_v3 && \
    yum clean all && \
    rm -rf /var/cache/yum

ENV KUSTOMIZE_VERSION v4.5.5

ENV KUSTOMIZE_URL=https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_${ARCH}.tar.gz

RUN curl -sLf ${KUSTOMIZE_URL} | tar -xzf - -C /usr/bin && \
    chmod +x /usr/bin/kustomize

WORKDIR /var/lib/rancher

ARG IMAGE_REPO=rancher
ARG SYSTEM_CHART_DEFAULT_BRANCH=release-v2.7
ARG PARTNER_CHART_DEFAULT_BRANCH=main
ARG RKE2_CHART_DEFAULT_BRANCH=main
# kontainer-driver-metadata branch to be set for specific branch other than dev/master, logic at rancher/rancher/pkg/settings/setting.go
ARG CATTLE_KDM_BRANCH=release-v2.7
ARG CATTLE_DASHBOARD_TAR_URL=https://github.com/verrazzano/rancher-dashboard/releases/download/v2.6.5-vz/rancher-dashboard-v2.6.5.tar.gz

ENV CATTLE_SYSTEM_CHART_DEFAULT_BRANCH=$SYSTEM_CHART_DEFAULT_BRANCH
ENV CATTLE_CHART_DEFAULT_BRANCH=oracle/release/2.7.2
ENV CATTLE_PARTNER_CHART_DEFAULT_BRANCH=$PARTNER_CHART_DEFAULT_BRANCH
ENV CATTLE_RKE2_CHART_DEFAULT_BRANCH=$RKE2_CHART_DEFAULT_BRANCH
ENV CATTLE_MACHINE_VERSION v0.15.0-rancher95
ENV CATTLE_MACHINE_PROVISION_IMAGE rancher/machine:${CATTLE_MACHINE_VERSION}
ENV CATTLE_ETCD_VERSION v3.5.1
ENV LOGLEVEL_VERSION v0.1.5
ENV TINI_VERSION v0.19.0
ENV TELEMETRY_VERSION v0.5.20
# make sure the version number is consistent with the one at Line 100 of pkg/data/management/machinedriver_data.go
ENV CATTLE_KDM_BRANCH ${CATTLE_KDM_BRANCH}
ENV HELM_VERSION 3.9.0-rancher1


# System charts minimal version
ENV CATTLE_FLEET_MIN_VERSION=102.0.0+up0.6.0
ENV CATTLE_RANCHER_WEBHOOK_MIN_VERSION=2.0.2+up0.3.2
ENV CATTLE_CSP_ADAPTER_MIN_VERSION=2.0.1

# Chart git repo directories
ENV RANCHER_CHARTS_DIR=/var/lib/rancher-data/local-catalogs/v2/rancher-charts/4b40cac650031b74776e87c1a726b0484d0877c3ec137da0872547ff9b73a721

# System chart commits
ENV RANCHER_CHARTS_COMMIT=7db2041af1d3b5a51167af967cf9c118b35efe31

RUN mkdir -p /var/lib/rancher-data/local-catalogs/system-library && \
    mkdir -p /var/lib/rancher-data/local-catalogs/library && \
    mkdir -p /var/lib/rancher-data/local-catalogs/helm3-library && \
    mkdir -p /var/lib/rancher-data/local-catalogs/v2 && \
    # Set up the git 2.x environment variables so that it's on the path
    source /etc/profile.d/git.sh && \
    #
    # Charts need to be copied into the sha256 value of git url computed in https://github.com/rancher/rancher/blob/5ebda9ac23c06e9647b586ec38aa51cc9ff9b031/pkg/catalogv2/git/download.go#L102 to create a unique folder per url
    #
    # After cloning the chart repo, checkout a specific commit so these builds are reproducible
    git clone -b $CATTLE_CHART_DEFAULT_BRANCH https://github.com/verrazzano/rancher-charts/ $RANCHER_CHARTS_DIR && \
    git -C $RANCHER_CHARTS_DIR checkout $RANCHER_CHARTS_COMMIT && \
    #
    git clone -b master --depth 1 https://github.com/verrazzano/rancher-charts/ /var/lib/rancher-data/local-catalogs/library && \
    git clone -b master --depth 1 https://github.com/rancher/helm3-charts /var/lib/rancher-data/local-catalogs/helm3-library


RUN curl -sLf https://github.com/rancher/telemetry/releases/download/${TELEMETRY_VERSION}/telemetry-${ARCH} > /usr/bin/telemetry && \
    chmod +x /usr/bin/tini /usr/bin/telemetry && \
    mkdir -p /var/lib/rancher-data/driver-metadata


# Versions for UI, DASHBOARD and CLI will be the same as 2.6 until new releases are made to these
ENV CATTLE_UI_VERSION 2.7.2
ENV CATTLE_DASHBOARD_UI_VERSION v2.7.2
ENV CATTLE_CLI_VERSION v2.7.0

# Base UI brand used as a fallback env setting (not user facing) to indicate this is a non-prime install
ENV CATTLE_BASE_UI_BRAND=

# Please update the api-ui-version in pkg/settings/settings.go when updating the version here.
ENV CATTLE_API_UI_VERSION 1.1.9

RUN mkdir -p /var/log/auditlog
ENV AUDIT_LOG_PATH /var/log/auditlog/rancher-api-audit.log
ENV AUDIT_LOG_MAXAGE 10
ENV AUDIT_LOG_MAXBACKUP 10
ENV AUDIT_LOG_MAXSIZE 100
ENV AUDIT_LEVEL 0

RUN mkdir -p /usr/share/rancher/ui && \
    cd /usr/share/rancher/ui && \
    curl -sL https://releases.rancher.com/ui/${CATTLE_UI_VERSION}.tar.gz | tar xvzf - --strip-components=1 && \
    mkdir -p /usr/share/rancher/ui/api-ui && \
    cd /usr/share/rancher/ui/api-ui && \
    curl -sL https://releases.rancher.com/api-ui/${CATTLE_API_UI_VERSION}.tar.gz | tar xvzf - --strip-components=1 && \
    mkdir -p /usr/share/rancher/ui-dashboard/dashboard && \
    cd /usr/share/rancher/ui-dashboard/dashboard && \
    curl -sL ${CATTLE_DASHBOARD_TAR_URL} | tar xvzf - --strip-components=2 && \
    ln -s dashboard/index.html ../index.html


ENV CATTLE_CLI_URL_DARWIN  https://releases.rancher.com/cli2/${CATTLE_CLI_VERSION}/rancher-darwin-amd64-${CATTLE_CLI_VERSION}.tar.gz
ENV CATTLE_CLI_URL_LINUX   https://releases.rancher.com/cli2/${CATTLE_CLI_VERSION}/rancher-linux-amd64-${CATTLE_CLI_VERSION}.tar.gz
ENV CATTLE_CLI_URL_WINDOWS https://releases.rancher.com/cli2/${CATTLE_CLI_VERSION}/rancher-windows-386-${CATTLE_CLI_VERSION}.zip

ARG VERSION=dev
ENV CATTLE_SERVER_VERSION ${VERSION}
COPY entrypoint.sh rancher /usr/bin/
COPY kustomize.sh /usr/bin/
COPY jailer.sh /usr/bin/
RUN mkdir /license
COPY LICENSE README.md THIRD_PARTY_LICENSES.txt /license/
RUN chmod +x /usr/bin/entrypoint.sh
RUN chmod +x /usr/bin/kustomize.sh

COPY data.json /var/lib/rancher-data/driver-metadata/

# Force use of rancher-agent image built from source for Verrazzano.
ENV CATTLE_AGENT_IMAGE ghcr.io/verrazzano/rancher-agent:${VERSION}
ENV CATTLE_SERVER_IMAGE ${IMAGE_REPO}/rancher
ENV ETCDCTL_API=3

ENV SSL_CERT_DIR /etc/rancher/ssl
VOLUME /var/lib/rancher
VOLUME /var/lib/kubelet
VOLUME /var/lib/cni
VOLUME /var/log

ENTRYPOINT ["entrypoint.sh"]
