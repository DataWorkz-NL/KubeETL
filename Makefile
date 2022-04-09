
# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/dataworkz-nl:main # TODO default to right approach (e.g. latest tag from github)
# CRD Options
CRD_OPTIONS ?= "crd"
KUBEBUILDER_ASSETS_DIR ?= "/usr/local/kubebuilder/bin"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	export KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS_DIR)
	go test -race ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Build hack binary
hack: fmt vet
	go build -o bin/hack ./hack

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go manager

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Set the image for the controller
.PHONY: set-image
set-image:
	cd config/manager && kustomize edit set image controller=${IMG}

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests set-image
	kustomize build config/overlays/default | kubectl apply -f -

# Generate quick-start yaml
.PHONY: quick-start set-image
quick-start: manifests
	kustomize build config/crd > manifests/quick-start.yaml
	kustomize build config/overlays/crd/webhook > manifests/quick-start-webhook.yaml
	kustomize build config/overlays/default >> manifests/quick-start.yaml
	kustomize build config/overlays/with_webhook >> manifests/quick-start-webhook.yaml

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen hack
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	bin/hack removecrdvalidation config/crd/bases/etl.dataworkz.nl_workflows.yaml

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.5.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
