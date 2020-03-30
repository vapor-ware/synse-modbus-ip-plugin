#
# Synse Modbus TCP/IP Plugin
#

PLUGIN_NAME    := modbus-ip
PLUGIN_VERSION := 1.1.2
IMAGE_NAME     := vaporio/modbus-ip-plugin

GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2> /dev/null || true)
GIT_TAG    ?= $(shell git describe --tags 2> /dev/null || true)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%T 2> /dev/null)
GO_VERSION := $(shell go version | awk '{ print $$3 }')

PKG_CTX := github.com/vapor-ware/synse-modbus-ip-plugin/vendor/github.com/vapor-ware/synse-sdk/sdk
LDFLAGS := -w \
	-X ${PKG_CTX}.BuildDate=${BUILD_DATE} \
	-X ${PKG_CTX}.GitCommit=${GIT_COMMIT} \
	-X ${PKG_CTX}.GitTag=${GIT_TAG} \
	-X ${PKG_CTX}.GoVersion=${GO_VERSION} \
	-X ${PKG_CTX}.PluginVersion=${PLUGIN_VERSION}


HAS_LINT := $(shell which golangci-lint)
HAS_GOX  := $(shell which gox)


#
# Local Targets
#

.PHONY: build
build:  ## Build the plugin Go binary
	go build -ldflags "${LDFLAGS}" -o build/plugin || exit

.PHONY: ci
ci:  ## Run CI checks locally (build, lint)
	@$(MAKE) build lint

.PHONY: clean
clean:  ## Remove temporary files
	go clean -v || exit

.PHONY: deploy
deploy:  ## Run a local deployment of Synse Server, and the Modbus TCP/IP plugin.
	docker-compose -f deploy/docker/deploy.yml up

.PHONY: docker
docker:  ## Build the docker image
	docker build -f Dockerfile \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg BUILD_VERSION=$(PLUGIN_VERSION) \
		--build-arg VCS_REF=$(GIT_COMMIT) \
		-t $(IMAGE_NAME):latest \
		-t $(IMAGE_NAME):local .

.PHONY: fmt
fmt:  ## Run goimports on all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do goimports -w "$$file" || exit ; done

.PHONY: github-tag
github-tag:  ## Create and push a tag with the current plugin version
	git tag -a ${PLUGIN_VERSION} -m "${PLUGIN_NAME} plugin version ${PLUGIN_VERSION}"
	git push -u origin ${PLUGIN_VERSION}

.PHONY: lint
lint:  ## Lint project source files
ifndef HAS_LINT
		$(shell curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $$(go env GOPATH)/bin v1.18.0)
endif
	golangci-lint run

.PHONY: setup
setup:  ## Install the build and development dependencies and set up vendoring
	$(shell curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $$(go env GOPATH)/bin v1.18.0)

.PHONY: test
test:  ## Run project tests
	go test -race -cover ./pkg/... || exit

.PHONY: version
version:  ## Print the version of the plugin
	@echo "$(PLUGIN_VERSION)"

.PHONY: help
help:  ## Print usage information
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.DEFAULT_GOAL := help


#
# CI Targets
#

.PHONY: ci-check-version
ci-check-version:
	PLUGIN_VERSION=$(PLUGIN_VERSION) ./bin/ci/check_version.sh

.PHONY: ci-build
ci-build:
ifndef HAS_GOX
	go get -v github.com/mitchellh/gox
endif
	@ # We currently only use a couple of images; the built set of images can be
	@ # updated if we ever need to support more os/arch combinations
	gox --output="build/${PLUGIN_NAME}_{{.OS}}_{{.Arch}}" \
		--ldflags "${LDFLAGS}" \
		--osarch='linux/amd64 darwin/amd64' \
		github.com/vapor-ware/synse-modbus-ip-plugin
