TARGETS := $(shell ls scripts)

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-`uname -s`-`uname -m` > .dapper.tmp
	@@chmod +x .dapper.tmp
	@./.dapper.tmp -v
	@mv .dapper.tmp .dapper

$(TARGETS): .dapper
	@if [ "$@" = "post-release-checks" ] || [ "$@" = "list-gomod-updates" ] || [ "$@" = "check-chart-kdm-source-values" ]; then \
		./.dapper -q --no-out --build-arg CATTLE_DASHBOARD_TAR_URL=${CATTLE_DASHBOARD_TAR_URL} $@; \
	else \
		./.dapper $@; \
	fi

.DEFAULT_GOAL := ci

.PHONY: $(TARGETS)
