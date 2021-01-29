PREFIX?=wavefronthq
DOCKER_IMAGE=wavefront-kubernetes-collector
ARCH?=amd64
OUT_DIR?=./_output
KUSTOMIZE_DIR=./hack/kustomize
GOLANG_VERSION?=1.13

BINARY_NAME=wavefront-collector

ifndef TEMP_DIR
TEMP_DIR:=$(shell mktemp -d /tmp/wavefront.XXXXXX)
endif

VERSION?=1.2.7
GIT_COMMIT:=$(shell git rev-parse --short HEAD)

REPO_DIR:=$(shell pwd)
KUSTOMIZE_DIR=${REPO_DIR}/hack/kustomize

# for testing, the built image will also be tagged with this name provided via an environment variable
OVERRIDE_IMAGE_NAME?=${COLLECTOR_TEST_IMAGE}

LDFLAGS=-w -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT)

all: build

fmt:
	find . -type f -name "*.go" | grep -v "./vendor*" | xargs gofmt -s -w

tests:
	go clean -testcache
	go test -timeout 30s -race ./...

build: clean fmt
	go vet -composites=false ./...
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(ARCH)/$(BINARY_NAME) ./cmd/wavefront-collector/

# test driver for local development
driver: clean fmt
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(ARCH)/$(BINARY_NAME)-test ./cmd/test-driver/

container:
	# Run build in a container in order to have reproducible builds
	docker run --rm -v $(TEMP_DIR):/build -v $(REPO_DIR):/go/src/github.com/wavefronthq/wavefront-collector-for-kubernetes -w /go/src/github.com/wavefronthq/wavefront-collector-for-kubernetes golang:$(GOLANG_VERSION) /bin/bash -c "\
		cp /etc/ssl/certs/ca-certificates.crt /build \
		&& GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags \"$(LDFLAGS)\" -o /build/$(BINARY_NAME) github.com/wavefronthq/wavefront-collector-for-kubernetes/cmd/wavefront-collector/"

	cp deploy/docker/Dockerfile $(TEMP_DIR)
	docker build --pull -t $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(TEMP_DIR)
	rm -rf $(TEMP_DIR)
ifneq ($(OVERRIDE_IMAGE_NAME),)
	docker tag $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(OVERRIDE_IMAGE_NAME)
endif

output-test:
	if [ -z ${WAVEFRONT_API_KEY} ]; then echo "Need to set WAVEFRONT_API_KEY" && exit 1; fi
	kind load docker-image wavefronthq/wavefront-kubernetes-collector:$(VERSION) --name kind
	(cd $(KUSTOMIZE_DIR) && ./test.sh nimba $(WAVEFRONT_API_KEY) $(VERSION))

#This rule need to be run on RHEL with podman installed.
container_rhel: build
	cp $(OUT_DIR)/$(ARCH)/$(BINARY_NAME) $(TEMP_DIR)
	cp LICENSE $(TEMP_DIR)/license.txt
	cp deploy/docker/Dockerfile-rhel $(TEMP_DIR)/Dockerfile
	cp deploy/examples/openshift-config.yaml $(TEMP_DIR)/collector.yaml
	sudo docker build --pull -t $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(TEMP_DIR)
	rm -rf $(TEMP_DIR)
ifneq ($(OVERRIDE_IMAGE_NAME),)
	sudo docker tag $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(OVERRIDE_IMAGE_NAME)
endif

clean:
	rm -f $(OUT_DIR)/$(ARCH)/$(BINARY_NAME)
	rm -f $(OUT_DIR)/$(ARCH)/$(BINARY_NAME)-test

.PHONY: all fmt container clean
