#
# Synse Modbus TCP/IP Plugin
#

PLUGIN_NAME    := modbus-ip
PLUGIN_VERSION := 2.1.5
IMAGE_NAME     := vaporio/modbus-ip-plugin
BIN_NAME       := synse-modbus-ip-plugin

GIT_COMMIT     ?= $(shell git rev-parse --short HEAD 2> /dev/null || true)
GIT_TAG        ?= $(shell git describe --tags 2> /dev/null || true)
BUILD_DATE     := $(shell date -u +%Y-%m-%dT%T 2> /dev/null)
GO_VERSION     := $(shell go version | awk '{ print $$3 }')

PKG_CTX := github.com/vapor-ware/synse-sdk/sdk
LDFLAGS := -w \
	-X ${PKG_CTX}.BuildDate=${BUILD_DATE} \
	-X ${PKG_CTX}.GitCommit=${GIT_COMMIT} \
	-X ${PKG_CTX}.GitTag=${GIT_TAG} \
	-X ${PKG_CTX}.GoVersion=${GO_VERSION} \
	-X ${PKG_CTX}.PluginVersion=${PLUGIN_VERSION}


.PHONY: build
build:  ## Build the plugin binary
	go build -ldflags "${LDFLAGS}" -o ${BIN_NAME}

.PHONY: build-linux
build-linux:  ## Build the plugin binarry for linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BIN_NAME} .

.PHONY: clean
clean:  ## Remove temporary files
	go clean -v || exit
	rm -rf dist

.PHONY: cover
cover: test ## Run tests and open the coverage report
	go tool cover -html=coverage.out || exit

.PHONY: dep
dep:  ## Verify and tidy gomod dependencies
	go mod verify || exit
	go mod tidy || exit

.PHONY: deploy
deploy:  ## Run a local deployment of the plugin with Synse Server
	docker-compose up -d || exit

.PHONY: docker
docker:  ## Build the production docker image locally
	docker build -f Dockerfile \
		--label "org.label-schema.build-date=${BUILD_DATE}" \
		--label "org.label-schema.vcs-ref=${GIT_COMMIT}" \
		--label "org.label-schema.version=${PLUGIN_VERSION}" \
		-t ${IMAGE_NAME}:latest . || exit

.PHONY: docker-dev
docker-dev:  ## Build the development docker image locally
	docker build -f Dockerfile.dev -t ${IMAGE_NAME}:dev-${GIT_COMMIT} . || exit

.PHONY: fmt
fmt:  ## Run goimports on all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do goimports -w "$$file" || exit ; done

.PHONY: github-tag
github-tag:  ## Create and push a tag with the current plugin version
	git tag -a ${PLUGIN_VERSION} -m "${PLUGIN_NAME} plugin version ${PLUGIN_VERSION}"
	git push -u origin ${PLUGIN_VERSION}

.PHONY: lint
lint:  ## Lint project source files
	golint -set_exit_status ./pkg/... || exit

.PHONY: test
test:  ## Run project tests
	@ # Note: this requires go1.10+ in order to do multi-package coverage reports
	go test -race -coverprofile=coverage.out -covermode=atomic ./pkg/... || exit

test-endpoints:  ## Run endpoint tests against the emulator from your dev box.
	# Start emulator
	docker-compose -f emulator/modbus/docker-compose.yaml up -d --build
	# Tests are driven by the local box (not by a container).
	# Hold on to the test exit code. Always teardown the emulator.
	go test -v ./test/... ; rc=$$? ; docker-compose -f emulator/modbus/docker-compose.yaml down ; exit $$rc

.PHONY: version
version:  ## Print the version of the plugin
	@echo "${PLUGIN_VERSION}"

.PHONY: help
help:  ## Print usage information
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.DEFAULT_GOAL := help


# Jenkins CI Targets

.PHONY: unit-test
unit-test: test

.PHONY: integration-test
integration-test:
	go test -v ./test/...
