# Image URL to use all building/pushing image targets
#IMG ?= ghcr.io/kluster-manager/license-proxyserver-addon:latest
IMG ?= docker.io/rokibulhasan114/license-proxyserver-addon:latest
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

##@ Build
.PHONY: build
build: fmt ## Build manager binary.
	GOFLAGS="" CGO_ENABLED=0 go build -o bin/license-proxyserver-addon cmd/main.go

.PHONY: run
run: fmt ## Run a controller from your host.
	go run cmd/main.go

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build --no-cache -t ${IMG} .

.PHONY: docker-push
docker-push: docker-build ## Build and Push docker image with the manager.
	docker push ${IMG}

.PHONY: docker-push-to-kind
docker-push-to-kind: docker-build ## Build and Push docker image with the manager.
	kind load docker-image ${IMG} --name=hub

.PHONY: deploy-crd
deploy-crd: ## Apply flux config crd
	cd api/ && make manifests
	kustomize build api/config/crd/ | kubectl apply -f -

.PHONY: deploy-helm
deploy-helm:
	make docker-push
	make undeploy-helm --ignore-errors
	kubectl wait --for=delete namespace/license-proxyserver-addon --timeout=300s
	make deploy-crd --ignore-errors
	cd deploy/helm/license-proxyserver-addon-manager && helm install license-proxyserver-addon-manager . --namespace open-cluster-management --create-namespace

.PHONY: deploy-to-kind
deploy-to-kind:
	make docker-push-to-kind
	make undeploy-helm --ignore-errors
	kubectl wait --for=delete namespace/license-proxyserver-addon --timeout=300s
	make deploy-crd --ignore-errors
	cd deploy/helm/license-proxyserver-addon-manager && helm install license-proxyserver-addon-manager . --namespace open-cluster-management --create-namespace

.PHONY: undeploy-helm
undeploy-helm:
	helm uninstall license-proxyserver-addon-manager -n open-cluster-management
