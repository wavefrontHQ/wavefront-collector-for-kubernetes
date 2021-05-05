PREFIX?=wavefronthq
GCP_PROJECT=wavefront-gcp-dev
DOCKER_IMAGE=wavefront-kubernetes-collector
ARCH?=amd64

REPO_DIR=$(shell git rev-parse --show-toplevel)
KUSTOMIZE_DIR=$(REPO_DIR)/hack/kustomize
DEPLOY_DIR=$(REPO_DIR)/hack/deploy
OUT_DIR?=$(REPO_DIR)/_output

GOLANG_VERSION?=1.15
BINARY_NAME=wavefront-collector


COLLECTOR_NAMESPACE="1-3-6-wavefront-collector"
TARGETS_NAMESPACE=collector-targets

ifndef TEMP_DIR
TEMP_DIR:=$(shell mktemp -d /tmp/wavefront.XXXXXX)
endif

VERSION?=1.3.6
GIT_COMMIT:=$(shell git rev-parse --short HEAD)

# for testing, the built image will also be tagged with this name provided via an environment variable
OVERRIDE_IMAGE_NAME?=${COLLECTOR_TEST_IMAGE}

LDFLAGS=-w -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT)

all: container

fmt:
	find . -type f -name "*.go" | grep -v "./vendor*" | xargs gofmt -s -w

tests:
	go clean -testcache
	go test -timeout 30s -race ./...

build: clean fmt vet
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(ARCH)/$(BINARY_NAME) ./cmd/wavefront-collector/

vet:
	go vet -composites=false ./...

# test driver for local development
driver: clean fmt
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(ARCH)/$(BINARY_NAME)-test ./cmd/test-driver/

nuke-loop: token-check nuke-kind deploy-targets full-loop

full-loop: token-check build tests containers output-test

nuke-kind:
	kind delete cluster
	kind create cluster

containers: container test-proxy-container

container:
	# Run build in a container in order to have reproducible builds
	docker run --rm -v $(TEMP_DIR):/build -v $(REPO_DIR):/go/src/github.com/wavefronthq/wavefront-collector-for-kubernetes -w /go/src/github.com/wavefronthq/wavefront-collector-for-kubernetes golang:$(GOLANG_VERSION) /bin/bash -c "\
		cp /etc/ssl/certs/ca-certificates.crt /build \
		&& GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags \"$(LDFLAGS)\" -o /build/$(BINARY_NAME) github.com/wavefronthq/wavefront-collector-for-kubernetes/cmd/wavefront-collector/"

	cp deploy/docker/Dockerfile $(TEMP_DIR)
	docker build --pull -t $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(TEMP_DIR)
ifneq ($(OVERRIDE_IMAGE_NAME),)
	docker tag $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(OVERRIDE_IMAGE_NAME)
endif

test-proxy-container:
	docker run --rm \
		-v $(REPO_DIR):/go/src/github.com/wavefronthq/wavefront-collector-for-kubernetes \
		-v $(TEMP_DIR):/go/src/github.com/wavefronthq/wavefront-collector-for-kubernetes/_output/$(ARCH) \
		-w /go/src/github.com/wavefronthq/wavefront-collector-for-kubernetes golang:$(GOLANG_VERSION) \
		/usr/bin/make test-proxy
	docker build --pull -f $(REPO_DIR)/hack/deploy/Dockerfile.test-proxy -t $(PREFIX)/test-proxy:$(VERSION) $(TEMP_DIR)

test-proxy: peg $(REPO_DIR)/cmd/test-proxy/metric_grammar.peg.go clean fmt vet
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(ARCH)/test-proxy ./cmd/test-proxy/...

peg:
	@which peg > /dev/null || \
		(cd $(REPO_DIR)/..; GOARCH=$(ARCH) CGO_ENABLED=0 go get -u github.com/pointlander/peg)

%.peg.go: %.peg
	peg -switch -inline $<

redeploy: token-check
	(cd $(KUSTOMIZE_DIR) && ./deploy.sh -c nimba -t ${WAVEFRONT_API_KEY} -v ${VERSION} -i "$(PREFIX)\/$(DOCKER_IMAGE)")

deploy-targets:
	(cd $(DEPLOY_DIR) && ./deploy-targets.sh)

output-test: token-check
	docker exec -it kind-control-plane crictl rmi $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) || true
	kind load docker-image $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) --name kind

	docker exec -it kind-control-plane crictl rmi $(PREFIX)/test-proxy:$(VERSION) || true
	kind load docker-image $(PREFIX)/test-proxy:$(VERSION) --name kind

	(cd $(KUSTOMIZE_DIR) && ./test.sh nimba $(WAVEFRONT_API_KEY) $(VERSION))

token-check:
	if [ -z ${WAVEFRONT_API_KEY} ]; then echo "Need to set WAVEFRONT_API_KEY" && exit 1; fi

k9s:
	watch -n 1 k9s

#This rule need to be run on RHEL with podman installed.
container_rhel: build
	cp $(OUT_DIR)/$(ARCH)/$(BINARY_NAME) $(TEMP_DIR)
	cp LICENSE $(TEMP_DIR)/license.txt
	cp deploy/docker/Dockerfile-rhel $(TEMP_DIR)/Dockerfile
	cp deploy/examples/openshift-config.yaml $(TEMP_DIR)/collector.yaml
	sudo docker build --pull -t $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(TEMP_DIR)
ifneq ($(OVERRIDE_IMAGE_NAME),)
	sudo docker tag $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(OVERRIDE_IMAGE_NAME)
endif

clean:
	rm -f $(OUT_DIR)/$(ARCH)/$(BINARY_NAME)
	rm -f $(OUT_DIR)/$(ARCH)/$(BINARY_NAME)-test
	rm -f $(OUT_DIR)/$(ARCH)/test-proxy

target-gke:
	gcloud config set project $(GCP_PROJECT)

nuke-gke: target-gke
# 	gcloud container clusters delete $(USER)-test-cluster-1 --region=us-central1-c --quiet
	#gcloud container clusters create $(USER)-test-cluster-1 --region=us-central1-c --enable-ip-alias --create-subnetwork range=/21
	kubectl create clusterrolebinding --clusterrole cluster-admin \
		--user $$(gcloud auth list --filter=status:ACTIVE --format="value(account)") \
		clusterrolebinding

clean-gke: target-gke
	kubectl delete namespace $(COLLECTOR_NAMESPACE) || true
	kubectl delete namespace $(TARGETS_NAMESPACE) || true

push-to-gcr: test-proxy-container container
	#docker build --pull -f $(REPO_DIR)/hack/deploy/Dockerfile.test-proxy -t $(PREFIX)/test-proxy:$(VERSION) $(TEMP_DIR)
	docker tag $(PREFIX)/test-proxy:$(VERSION) us.gcr.io/$(GCP_PROJECT)/test-proxy:$(VERSION)
	docker push us.gcr.io/$(GCP_PROJECT)/test-proxy:$(VERSION)

	docker tag $(PREFIX)/wavefront-kubernetes-collector:$(VERSION) us.gcr.io/$(GCP_PROJECT)/wavefront-kubernetes-collector:$(VERSION)
	docker push us.gcr.io/$(GCP_PROJECT)/wavefront-kubernetes-collector:$(VERSION)

output-test-gke: token-check
	(cd $(KUSTOMIZE_DIR) && ./test.sh nimba $(WAVEFRONT_API_KEY) $(VERSION) "us.gcr.io\/$(GCP_PROJECT)")

.PHONY: all fmt container clean
