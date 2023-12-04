TARGETS := $(shell ls scripts)

GO ?= CGO_ENABLED=0 GO111MODULE=on go
DAPPER_VERSION = v0.6.0-v8o-test

RANCHER_BASE_TAG ?= base
RANCHER_BASE_REPO ?= rancher

# find or download dapper
DAPPER_PATH := $(shell eval go env GOPATH)
.PHONY: dapper
dapper:
ifeq (, $(shell command -v dapper))
	$(GO) install github.com/verrazzano/rancher-dapper@${DAPPER_VERSION}
	mv ${DAPPER_PATH}/bin/rancher-dapper $(DAPPER_PATH)/bin/dapper
	$(eval DAPPER=$(DAPPER_PATH)/bin/dapper)
else
	$(eval DAPPER=$(shell command -v dapper))
endif

$(TARGETS): dapper
	docker build -f Dockerfile.base -t ${RANCHER_BASE_REPO}/rancher:${RANCHER_BASE_TAG} .
	@if [[ "$@" = "post-release-checks" ]] || [[ "$@" = "list-gomod-updates" ]] || [[ "$@" = "check-chart-kdm-source-values" ]]; then\
		dapper -q --no-out $@;\
	else\
		dapper $@;\
	fi

.DEFAULT_GOAL := ci

.PHONY: $(TARGETS)
