include project.mk
include boilerplate/generated-includes.mk

SHELL := /usr/bin/env bash

# Verbosity
AT_ = @
AT = $(AT_$(V))
# /Verbosity

GIT_HASH := $(shell git rev-parse --short=7 HEAD)

BINARY_FILE ?= build/_output/compliance-audit-router

GO_SOURCES := $(find $(CURDIR) -type f -name "*.go" -print)
EXTRA_DEPS := $(find $(CURDIR)/build -type f -print) Makefile

# Containers may default GOFLAGS=-mod=vendor which would break us since
# we're using modules.
unexport GOFLAGS
GOOS?=linux
GOARCH?=amd64
GOENV=GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 GOFLAGS=

GOBUILDFLAGS=-gcflags="all=-trimpath=${GOPATH}" -asmflags="all=-trimpath=${GOPATH}"

TESTOPTS ?=-v

.PHONY: test
test: vet $(GO_SOURCES)
	$(AT)go test $(TESTOPTS) $(shell go list -mod=readonly -e ./...)
	
.PHONY: clean
clean:
	$(AT)rm -f $(BINARY_FILE) 

.PHONY: serve
serve:
	$(AT)go run ./cmd/main.go 

.PHONY: vet
vet:
	$(AT)gofmt -s -l $(shell go list -f '{{ .Dir }}' ./... ) | grep ".*\.go"; if [ "$$?" = "0" ]; then gofmt -s -d $(shell go list -f '{{ .Dir }}' ./... ); exit 1; fi
	$(AT)go vet ./cmd/... ./pkg/...

.PHONY: build
build: $(BINARY_FILE)

$(BINARY_FILE): $(GO_SOURCES)
	mkdir -p $(shell dirname $(BINARY_FILE))
	$(GOENV) go build $(GOBUILDFLAGS) -o $(BINARY_FILE) ./cmd

# This is a helper for local development, not to be used for CI/CD
.PHONY: build-image
build-image: isclean
	${CONTAINER_ENGINE} build -f $(DOCKERFILE) -t $(IMAGE_NAME):$(IMAGE_TAG) .
	${CONTAINER_ENGINE} tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest

.PHONY: boilerplate-update
boilerplate-update:
	@boilerplate/update

# Just a helper to list all available make targets
.PHONY: list
list:
	@LC_ALL=C $(MAKE) -pRrq -f $(firstword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/(^|\n)# Files(\n|$$)/,/(^|\n)# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | grep -E -v -e '^[^[:alnum:]]' -e '^$@$$'
