MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash
.SHELLFLAGS := -o pipefail -euc
.DEFAULT_GOAL := build

.PHONY: clean test integration consul ship dockerfile docker cover lint local vendor dep-* tools kirby

IMPORT_PATH := github.com/asokolov365/containerpilot
VERSION ?= dev-build-not-for-release
LDFLAGS := -X ${IMPORT_PATH}/version.GitHash=$(shell git rev-parse --short HEAD) -X ${IMPORT_PATH}/version.Version=${VERSION}

ROOT := $(shell pwd)
RUNNER := -v ${ROOT}:/go/src/${IMPORT_PATH} -w /go/src/${IMPORT_PATH} containerpilot_build
docker := docker run --rm -e LDFLAGS="${LDFLAGS}" $(RUNNER)
date := $(shell date)
export PATH :=$(PATH):$(GOPATH)/bin

# flags for local development
GOPATH ?= $(shell go env GOPATH)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED := 0
GOEXPERIMENT := framepointer

CONSUL_VERSION := 1.9.4

## display this help message
help:
	@echo -e "\033[32m"
	@echo "Targets in this Makefile build and test ContainerPilot in a build container in"
	@echo "Docker. For testing (only), use the 'local' prefix target to run targets directly"
	@echo "on your workstation (ex. 'make local test'). You will need to have its GOPATH set"
	@echo "and have already run 'make tools'. Set GOOS=linux to build binaries for Docker."
	@echo "Do not use 'make local' for building binaries for public release!"
	@echo "Before packaging always run 'make clean build test integration'!"
	@echo
	@awk '/^##.*$$/,/[a-zA-Z_-]+:/' $(MAKEFILE_LIST) | awk '!(NR%2){print $$0p}{p=$$0}' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' | sort


# ----------------------------------------------
# building

## build the ContainerPilot binary
build: build/containerpilot
build/containerpilot:  build/containerpilot_build build/deps-installed */*/*.go */*.go */*/*.go *.go
	$(docker) go build -o build/containerpilot -ldflags "$(LDFLAGS)"
	@rm -rf src || true

# builds the builder container
build/containerpilot_build:
	mkdir -p ${ROOT}/build
	docker rmi -f containerpilot_build > /dev/null 2>&1 || true
	docker build -t containerpilot_build ${ROOT}
	docker inspect -f "{{ .ID }}" containerpilot_build > build/containerpilot_build

# Before packaging always `make clean build test integration`!
## tag and package ContainerPilot for release; `VERSION=make release`
release: build
	mkdir -p release
	git tag $(VERSION)
	git push --tags
	cd build && tar -cfz ../release/containerpilot-$(VERSION).tar.gz containerpilot
	@echo
	@cd release && shasum containerpilot-$(VERSION).tar.gz
	@cd release && shasum containerpilot-$(VERSION).tar.gz > containerpilot-$(VERSION).sha1.txt
	@echo Upload files in release/ directory to GitHub release.

## remove build/test artifacts, test fixtures, and vendor directories
clean:
	rm -rf build release cover vendor
	docker rmi -f containerpilot_build > /dev/null 2>&1 || true
	./scripts/test.sh clean

# ----------------------------------------------
# dependencies
## install any changed packages in the go.mod
vendor: build/deps-installed
build/deps-installed: build/containerpilot_build go.mod
	mkdir -p vendor
	mkdir -p ${ROOT}/build
	$(docker) go mod vendor -v
	@echo $(date) > build/deps-installed

## install all vendored packages in the go.mod
dep-install:
	mkdir -p vendor
	mkdir -p ${ROOT}/build
	$(docker) go mod vendor -v
	@echo $(date) > build/deps-installed

# run 'GOOS=darwin make tools' if you're installing on MacOS
## set up local dev environment
tools:
	@go version | grep 1.16 || (echo 'WARNING: go1.16 should be installed!')
	@$(if $(value GOPATH),, $(error 'GOPATH not set'))
	go get -u golang.org/x/lint/golint
	go get -u golang.org/x/tools/cmd/stringer
	curl --fail -Lso consul.zip "https://releases.hashicorp.com/consul/$(CONSUL_VERSION)/consul_$(CONSUL_VERSION)_$(GOOS)_$(GOARCH).zip"
	unzip consul.zip -d "$(GOPATH)/bin"
	rm consul.zip

# ----------------------------------------------
# develop and test

## print environment info about this dev environment
env:
	@$(if $(value DOCKER_HOST), echo "DOCKER_HOST=$(DOCKER_HOST)", echo 'DOCKER_HOST not set')
	@echo CGO_ENABLED=$(CGO_ENABLED)
	@echo GOARCH=$(GOARCH)
	@echo GOEXPERIMENT=$(GOEXPERIMENT)
	@echo GOOS=$(GOOS)
	@echo GOPATH=$(GOPATH)
	@echo IMPORT_PATH=$(IMPORT_PATH)
	@echo LDFLAGS="$(LDFLAGS)"
	@echo PATH=$(PATH)
	@echo ROOT=$(ROOT)
	@echo VERSION=$(VERSION)
	@echo
	@echo docker commands run as:
	@echo $(docker)

## prefix before other make targets to run in your local dev environment
local: | quiet
	@$(eval docker= )
quiet: # this is silly but shuts up 'Nothing to be done for `local`'
	@:

## run `go lint` and other code quality tools
lint: build/containerpilot_build
	$(docker) bash ./scripts/lint.sh

## run unit tests
test: build/containerpilot_build
	$(docker) bash ./scripts/unit_test.sh

## run unit tests and write out HTML file of test coverage
cover: build/containerpilot_build
	mkdir -p cover
	$(docker) bash ./scripts/cover.sh


## generate stringer code
generate:
	go install github.com/asokolov365/containerpilot/events
	cd events && stringer -type EventCode
	# fix this up for making it pass linting
	sed -i '.bak' 's/_EventCode_/eventCode/g' ./events/eventcode_string.go
	@rm -f ./events/eventcode_string.go.bak

TEST ?= "all"
## run integration tests; filter with `TEST=testname make integration`
integration: build
	./scripts/test.sh test $(TEST)


## build documentation for Kirby
kirby: build/docs

## preview the Kirby documentation
kirby-preview: build/docs
	docker run --rm -it -p 80:80 \
		-v ${ROOT}/build/docs:/var/www/html/content/1-containerpilot/1-docs/ \
		joyent/kirby-preview-base:latest

build/docs: docs/* scripts/docs.py
	rm -rf build/docs
	./scripts/docs.py
