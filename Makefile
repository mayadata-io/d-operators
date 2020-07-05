PWD := ${CURDIR}

OS = $(shell uname)

GIT_TAGS = $(shell git fetch --all --tags)
PACKAGE_VERSION ?= $(shell git describe --always --tags)
ALL_SRC = $(shell find . -name "*.go" | grep -v -e "vendor")
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# We are using docker hub as the default registry
#IMG_REGISTRY ?= quay.io
IMG_NAME ?= dope
IMG_REPO ?= mayadataio/dope

all: bins manifests

bins: vendor $(IMG_NAME)

$(IMG_NAME): $(ALL_SRC)
	@echo "+ Generating $(IMG_NAME) binary"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
		go build -o $@ ./cmd/main.go

$(ALL_SRC): ;

$(GIT_TAGS): ;

manifests: controller-gen

$(CONTROLLER_GEN) $(CRD_OPTIONS) paths="./types/jobs/..." output:crd:artifacts:config=manifests

# go mod download modules to local cache
# make vendored copy of dependencies
# install other go binaries for code generation
.PHONY: vendor
vendor: go.mod go.sum
	@GO111MODULE=on go mod download
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor

.PHONY: test
test: 
	@go test ./... -cover

.PHONY: testv
testv:
	@go test ./... -cover -v -args --logtostderr -v=2

.PHONY: image
image: $(GIT_TAGS)
	docker build -t $(IMG_REPO):$(PACKAGE_VERSION) .

.PHONY: push
push: image
	docker push $(IMG_REPO):$(PACKAGE_VERSION)