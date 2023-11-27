ARG RANCHER_BASE_TAG=base
ARG RANCHER_REPO=rancher

FROM ghcr.io/oracle/oraclelinux:8-slim as build
ARG ARCH
ENV ARCH=$ARCH

COPY . /go/src/github.com/rancher/rancher/

RUN microdnf update -y --setopt=install_weak_deps=0 --setopt=tsflags=nodocs && \
    microdnf install -y git && \
    microdnf install -y golang-1.20.10-1.module+el8.9.0+90068+08775a1f.x86_64 && \
    cd /go/src/github.com/rancher/rancher/scripts && \
    ./build

# Use the build of Dockerfile.base as the FROM image
FROM ${RANCHER_REPO}/rancher:${RANCHER_BASE_TAG}

ARG ARCH=amd64
ARG ETCD_UNSUPPORTED_ARCH

RUN useradd rancher && \
    rm -rf /tmp/* /var/tmp/* /usr/share/doc/packages/* && \
    mkdir -p /var/lib/rancher /var/lib/cattle /opt/jail /opt/drivers/management-state/bin && \
    chown -R rancher /var/lib/rancher /var/lib/cattle /usr/local/bin && \
    ln -s /usr/bin/rancher /usr/bin/reset-password && \
    ln -s /usr/bin/rancher /usr/bin/ensure-default-admin && \
    rm -f /bin/sh && ln -s /bin/bash /bin/sh

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
ARG CATTLE_UI_TAR_URL=https://github.com/verrazzano/rancher-ui/releases/download/v2.7.2-2/release-2.7.2.tar.gz
# Specific version of OCI OCNE Driver
ARG OCI_OCNE_DRIVER_VERSION=v0.23.0
ARG OCI_OCNE_DRIVER_HASH="e410fd9b7ac5533b656277e82f1a321a995a5cd5e0329c2e79091e26e2516313"
# Specific version of OKE Driver
ARG OKE_DRIVER_VERSION=v1.8.4
ARG OKE_DRIVER_HASH="dd971c8630f3e7d17dde21ef23cfc007862770a4c0ce21f3101929ff39961259"
# Specific version of OKE CAPI Driver
ARG OKE_CAPI_DRIVER_VERSION=v0.5.0
ARG OKE_CAPI_DRIVER_HASH="a2eb7a6f94c0136df2d13c5a5f6299873a1ca136046a999fd994444b411be33c"

ENV CATTLE_SYSTEM_CHART_DEFAULT_BRANCH=$SYSTEM_CHART_DEFAULT_BRANCH
ENV CATTLE_CHART_DEFAULT_BRANCH=release-v2.7
ENV CATTLE_PARTNER_CHART_DEFAULT_BRANCH=$PARTNER_CHART_DEFAULT_BRANCH
ENV CATTLE_RKE2_CHART_DEFAULT_BRANCH=$RKE2_CHART_DEFAULT_BRANCH
ENV CATTLE_MACHINE_VERSION v0.15.0-rancher100
ENV CATTLE_K3S_VERSION v1.26.4+k3s1
ENV CATTLE_MACHINE_PROVISION_IMAGE rancher/machine:${CATTLE_MACHINE_VERSION}
ENV CATTLE_ETCD_VERSION v3.5.1
ENV LOGLEVEL_VERSION v0.1.5
ENV TINI_VERSION v0.19.0
ENV TELEMETRY_VERSION v0.5.20
ENV DOCKER_MACHINE_LINODE_VERSION v0.1.8
ENV LINODE_UI_DRIVER_VERSION v0.5.0
# make sure the version number is consistent with the one at Line 100 of pkg/data/management/machinedriver_data.go
ENV CATTLE_KDM_BRANCH ${CATTLE_KDM_BRANCH}
ENV OCI_OCNE_DRIVER_VERSION=$OCI_OCNE_DRIVER_VERSION
ENV OCI_OCNE_DRIVER_HASH=$OCI_OCNE_DRIVER_HASH
ENV OCI_OCNE_DRIVER_PATH=/var/lib/rancher-data/drivers/ociocne/$OCI_OCNE_DRIVER_VERSION
ENV OCI_OCNE_DRIVER_BINARY=kontainer-engine-driver-ociocne-linux
ENV OKE_DRIVER_VERSION=$OKE_DRIVER_VERSION
ENV OKE_DRIVER_HASH=$OKE_DRIVER_HASH
ENV OKE_DRIVER_PATH=/var/lib/rancher-data/drivers/oke/$OKE_DRIVER_VERSION
ENV OKE_DRIVER_BINARY=kontainer-engine-driver-oke-linux
ENV OKE_CAPI_DRIVER_VERSION=$OKE_CAPI_DRIVER_VERSION
ENV OKE_CAPI_DRIVER_HASH=$OKE_CAPI_DRIVER_HASH
ENV OKE_CAPI_DRIVER_PATH=/var/lib/rancher-data/drivers/okecapi/$OKE_CAPI_DRIVER_VERSION
ENV OKE_CAPI_DRIVER_BINARY=kontainer-engine-driver-okecapi-linux

# System charts minimal version
ENV CATTLE_FLEET_MIN_VERSION=102.1.1+up0.7.1
# Deprecated in favor of CATTLE_RANCHER_WEBHOOK_VERSION.
ENV CATTLE_RANCHER_WEBHOOK_MIN_VERSION=''
ENV CATTLE_RANCHER_WEBHOOK_VERSION=2.0.5+up0.3.5
ENV CATTLE_CSP_ADAPTER_MIN_VERSION=2.0.2

# Chart git repo directories
ENV RANCHER_CHARTS_DIR=/var/lib/rancher-data/local-catalogs/v2/rancher-charts/4b40cac650031b74776e87c1a726b0484d0877c3ec137da0872547ff9b73a721

# System chart commits
ENV RANCHER_CHARTS_COMMIT=b573e1310cd9cb7716650928e81d73f9222f4be6

RUN mkdir -p /var/lib/rancher-data/local-catalogs/system-library && \
    mkdir -p /var/lib/rancher-data/local-catalogs/library && \
    mkdir -p /var/lib/rancher-data/local-catalogs/helm3-library && \
    mkdir -p /var/lib/rancher-data/local-catalogs/v2 && \
    mkdir -p $OCI_OCNE_DRIVER_PATH  && \
    mkdir -p $OKE_DRIVER_PATH  && \
    mkdir -p $OKE_CAPI_DRIVER_PATH  && \
    #
    # Charts need to be copied into the sha256 value of git url computed in https://github.com/rancher/rancher/blob/5ebda9ac23c06e9647b586ec38aa51cc9ff9b031/pkg/catalogv2/git/download.go#L102 to create a unique folder per url
    #
    # After cloning the chart repo, checkout a specific commit so these builds are reproducible
    git clone -b $CATTLE_CHART_DEFAULT_BRANCH https://git.rancher.io/charts/ $RANCHER_CHARTS_DIR && \
    git -C $RANCHER_CHARTS_DIR checkout $RANCHER_CHARTS_COMMIT && \
    # Remove charts that are not used \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-alerting-drivers && \
    rm -rf $RANCHER_CHARTS_DIR/assets/rancher-alerting-drivers && \
    rm -rf $RANCHER_CHARTS_DIR/packages/rancher-alerting/rancher-alerting-drivers && \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-cis-benchmark && \
    rm -rf $RANCHER_CHARTS_DIR/assets/rancher-cis-benchmark && \
    rm -rf $RANCHER_CHARTS_DIR/packages/rancher-cis-benchmark && \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-cis-benchmark-crd && \
    rm -rf $RANCHER_CHARTS_DIR/assets/rancher-cis-benchmark-crd && \
    rm -rf $RANCHER_CHARTS_DIR/packages/rancher-cis-benchmark-crd && \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-istio && \
    rm -rf $RANCHER_CHARTS_DIR/assets/rancher-istio && \
    rm -rf $RANCHER_CHARTS_DIR/packages/rancher-istio && \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-pushprox && \
    rm -rf $RANCHER_CHARTS_DIR/assets/rancher-pushprox && \
    rm -rf $RANCHER_CHARTS_DIR/packages/rancher-monitoring/rancher-pushprox && \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-monitoring/9.4.* && \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-sachet && \
    rm -rf $RANCHER_CHARTS_DIR/assets/rancher-sachet && \
    rm -rf $RANCHER_CHARTS_DIR/packages/rancher-alerting/rancher-sachet && \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-tracing && \
    rm -rf $RANCHER_CHARTS_DIR/assets/rancher-tracing && \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-windows-exporter && \
    rm -rf $RANCHER_CHARTS_DIR/assets/rancher-windows-exporter && \
    rm -rf $RANCHER_CHARTS_DIR/packages/rancher-monitoring/rancher-windows-exporter && \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-windows-gmsa-crd && \
    rm -rf $RANCHER_CHARTS_DIR/assets/rancher-windows-gmsa-crd && \
    rm -rf $RANCHER_CHARTS_DIR/charts/rancher-windows-gmsa && \
    rm -rf $RANCHER_CHARTS_DIR/assets/rancher-windows-gmsa && \
    rm -rf $RANCHER_CHARTS_DIR/packages/rancher-windows-gmsa && \
    #
    git clone -b master --depth 1 https://github.com/rancher/charts/ /var/lib/rancher-data/local-catalogs/library && \
    git clone -b master --depth 1 https://github.com/rancher/helm3-charts /var/lib/rancher-data/local-catalogs/helm3-library && \
    # Download the OCI OCNE Driver
    curl -sLf https://github.com/verrazzano/kontainer-engine-driver-ociocne/releases/download/${OCI_OCNE_DRIVER_VERSION}/${OCI_OCNE_DRIVER_BINARY} > ${OCI_OCNE_DRIVER_PATH}/${OCI_OCNE_DRIVER_BINARY} && \
    # Download the OKE Driver
    curl -sLf https://github.com/rancher-plugins/kontainer-engine-driver-oke/releases/download/${OKE_DRIVER_VERSION}/${OKE_DRIVER_BINARY} > ${OKE_DRIVER_PATH}/${OKE_DRIVER_BINARY} &&  \
    # Download the OKE CAPI Driver
    curl -sLf https://github.com/verrazzano/kontainer-engine-driver-oke-capi/releases/download/${OKE_CAPI_DRIVER_VERSION}/${OKE_CAPI_DRIVER_BINARY} > ${OKE_CAPI_DRIVER_PATH}/${OKE_CAPI_DRIVER_BINARY}



RUN mkdir -p /var/lib/rancher/k3s/agent/images/ && \
    curl -sLf https://github.com/rancher/telemetry/releases/download/${TELEMETRY_VERSION}/telemetry-${ARCH} > /usr/bin/telemetry && \
    chmod +x /usr/bin/tini /usr/bin/telemetry && \
    mkdir -p /var/lib/rancher-data/driver-metadata

# Versions for UI, DASHBOARD and CLI will be the same as 2.6 until new releases are made to these
ENV CATTLE_UI_VERSION 2.7.6
ENV CATTLE_DASHBOARD_UI_VERSION v2.7.5
ENV CATTLE_CLI_VERSION v2.7.0

# Base UI brand used as a fallback env setting (not user facing) to indicate this is a non-prime install
ENV CATTLE_BASE_UI_BRAND=

# Please update the api-ui-version in pkg/settings/settings.go when updating the version here.
ENV CATTLE_API_UI_VERSION 1.1.10

RUN mkdir -p /var/log/auditlog
ENV AUDIT_LOG_PATH /var/log/auditlog/rancher-api-audit.log
ENV AUDIT_LOG_MAXAGE 10
ENV AUDIT_LOG_MAXBACKUP 10
ENV AUDIT_LOG_MAXSIZE 100
ENV AUDIT_LEVEL 0

RUN mkdir -p /usr/share/rancher/ui && \
    cd /usr/share/rancher/ui && \
    curl -sL ${CATTLE_UI_TAR_URL} | tar xvzf - --strip-components=1 && \
    mkdir -p assets/rancher-ui-driver-linode && \
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
COPY --from=build /go/src/github.com/rancher/rancher/package/entrypoint.sh /usr/bin/
COPY --from=build /go/src/github.com/rancher/rancher/bin/rancher /usr/bin/
COPY --from=build /go/src/github.com/rancher/rancher/package/kustomize.sh /usr/bin/
COPY --from=build /go/src/github.com/rancher/rancher/package/jailer.sh /usr/bin/
RUN mkdir /license
COPY --from=build /go/src/github.com/rancher/rancher/LICENSE /go/src/github.com/rancher/rancher/README.md /go/src/github.com/rancher/rancher/THIRD_PARTY_LICENSES.txt /license/
RUN chmod +x /usr/bin/entrypoint.sh
RUN chmod +x /usr/bin/kustomize.sh

COPY --from=build /go/src/github.com/rancher/rancher/package/data.json /var/lib/rancher-data/driver-metadata/

# Force use of rancher-agent image built from source for Verrazzano.
ENV CATTLE_AGENT_IMAGE ghcr.io/verrazzano/rancher-agent:${VERSION}
ENV CATTLE_SERVER_IMAGE ${IMAGE_REPO}/rancher
ENV ETCDCTL_API=3

ENV SSL_CERT_DIR /etc/rancher/ssl
VOLUME /var/lib/rancher
VOLUME /var/lib/kubelet
VOLUME /var/lib/cni
VOLUME /var/log

ENV ETCD_UNSUPPORTED_ARCH ${ETCD_UNSUPPORTED_ARCH}

ENTRYPOINT ["entrypoint.sh"]
