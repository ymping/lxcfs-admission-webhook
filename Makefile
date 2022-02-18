PROJECT_NAME := lxcfs-admission-webhook
#VERSION := $(shell git describe --tags)
COMMIT_ID := $(shell git rev-parse --short HEAD)

# Go related variables.
BASE_DIR := $(shell pwd)
BUILD_DIR := $(BASE_DIR)/build
WEBHOOK_BIN := $(BUILD_DIR)/$(PROJECT_NAME)
GO_COVERAGE := $(BUILD_DIR)/coverage.out
$(shell go env -w GOPROXY=https://goproxy.io,direct)
GO_MOD := $(shell go list -m)

# "go list $(GO_MOD)/..." need time to download dependence, run only necessary
NEED_GO_PKG_CMD := vet test test-coverage build run-wh
NEED_GO_PKG := $(foreach t,$(MAKECMDGOALS),$(filter $(t),$(NEED_GO_PKG_CMD)))
ifneq ($(NEED_GO_PKG),)
	GO_PKG := $(shell go list $(GO_MOD)/...)
endif

# Docker related variables.
DOCKER_USER := ymping
DOCKER_IMAGE_WH := $(DOCKER_USER)/$(PROJECT_NAME)
DOCKER_IMAGE_LXCFS := $(DOCKER_USER)/lxcfs

.PHONY: all dep lint vet test test-coverage build clean start-wh build-image-wh push-image-wh build-image-lxcfs push-image-lxcfs

all: help

dep: ## Get the dependencies
	@go mod download

lint: ## Lint Golang files
	@golangci-lint run

vet: ## Run go vet
	@go vet $(GO_PKG)

test: ## Run unittests
	@go test -short $(GO_PKG)

test-coverage: ## Run tests with coverage
	@go test -short -coverprofile=$(GO_COVERAGE) -covermode=atomic $(GO_PKG)

build: dep ## Build the binary file
	@go build -o $(WEBHOOK_BIN) $(GO_PKG)

clean: ## Remove previous build
	@-rm -f $(WEBHOOK_BIN)
	@-rm -f $(GO_COVERAGE)

run-wh: build ## Start lxcfs admission webhook
	@cd deploy; sh ./install.sh --create-cert-only
	@$(WEBHOOK_BIN) -tlsCertFile=$(BASE_DIR)/deploy/certs/server-cert.pem \
					-tlsKeyFile=$(BASE_DIR)/deploy/certs/server-key.pem \
					-alsologtostderr \
					-v=4 \
					2>&1

build-image-wh: ## Build lxcfs admission webhook docker images
	@docker build -t $(DOCKER_IMAGE_WH):$(COMMIT_ID) .

push-image-wh: build-image-wh ## Push lxcfs admission webhook docker images
	@docker push $(DOCKER_IMAGE_WH):$(COMMIT_ID)

build-image-lxcfs: ## Build lxcfs docker images, use example: make LXCFS_VERSION=4.0.11-r0 build-image-lxcfs
	@cd lxcfs-image; docker build -t $(DOCKER_IMAGE_LXCFS):$(LXCFS_VERSION) --build-arg LXCFS_VERSION=$(LXCFS_VERSION) .

push-image-lxcfs: build-image-lxcfs ## Build lxcfs docker images, use example: make LXCFS_VERSION=4.0.11-r0 push-image-lxcfs
	@docker push $(DOCKER_IMAGE_LXCFS):$(LXCFS_VERSION)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
