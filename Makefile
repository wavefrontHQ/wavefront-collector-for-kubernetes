PREFIX?=wavefronthq
DOCKER_IMAGE?=wavefront-kubernetes-collector
ARCH?=amd64

REPO_DIR=$(shell git rev-parse --show-toplevel)
KUSTOMIZE_DIR=$(REPO_DIR)/hack/kustomize
DEPLOY_DIR=$(REPO_DIR)/hack/deploy
OUT_DIR?=$(REPO_DIR)/_output

GOLANG_VERSION?=1.15
BINARY_NAME=wavefront-collector

RELEASE_TYPE?=release
RC_NUMBER?=1
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GIT_HUB_REPO=wavefrontHQ/wavefront-collector-for-kubernetes

K8S_ENV=$(shell cd $(DEPLOY_DIR) && ./get-k8s-cluster-env.sh)

ifndef TEMP_DIR
TEMP_DIR:=$(shell mktemp -d /tmp/wavefront.XXXXXX)
endif

GO_IMPORTS_BIN:=$(if $(which goimports),$(which goimports),$(GOPATH)/bin/goimports)
SEMVER_CLI_BIN:=$(if $(which semver-cli),$(which semver-cli),$(GOPATH)/bin/semver-cli)

VERSION_POSTFIX?=""
RELEASE_VERSION?=$(shell cat ./release/VERSION)
VERSION?=$(shell semver-cli inc patch $(RELEASE_VERSION))$(VERSION_POSTFIX)
GIT_COMMIT:=$(shell git rev-parse --short HEAD)

# for testing, the built image will also be tagged with this name provided via an environment variable
OVERRIDE_IMAGE_NAME?=${COLLECTOR_TEST_IMAGE}

LDFLAGS=-w -X main.version=$(RELEASE_VERSION) -X main.commit=$(GIT_COMMIT)

include make/k8s-envs/*.mk

all: container

fmt: $(GO_IMPORTS_BIN)
	find . -type f -name "*.go" | grep -v "./vendor*" | xargs goimports -w

checkfmt: $(GO_IMPORTS_BIN)
	@if [ $$(goimports -d $$(find . -type f -name '*.go' -not -path "./vendor/*") | wc -l) -gt 0 ]; then \
		echo $$'\e[31mgoimports FAILED!!!\e[0m'; \
		goimports -d $$(find . -type f -name '*.go' -not -path "./vendor/*"); \
		exit 1; \
	fi

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

containers: container test-proxy-container

container: #$(SEMVER_CLI_BIN)
	# Run build in a container in order to have reproducible builds
	docker build \
	--build-arg BINARY_NAME=$(BINARY_NAME) --build-arg LDFLAGS="$(LDFLAGS)" .

github-release:
	curl -X POST -H "Content-Type:application/json" -H "Authorization: token $(GITHUB_TOKEN)" \
		-d '{"tag_name":"v$(RELEASE_VERSION)", "target_commitish":"$(GIT_BRANCH)", "name":"Release v$(RELEASE_VERSION)", "body": "Description for v$(RELEASE_VERSION)", "draft": true, "prerelease": false}' "https://api.github.com/repos/$(GIT_HUB_REPO)/releases"

docker-login:
	echo '$(DOCKER_CREDS_PSW)' | docker login --username '$(DOCKER_CREDS_USR)' --password-stdin $(PREFIX)

publish: docker-login release

release:
	docker run --rm --privileged harbor-repo.vmware.com/dockerhub-proxy-cache/multiarch/qemu-user-static@sha256:c772ee1965aa0be9915ee1b018a0dd92ea361b4fa1bcab5bbc033517749b2af4 --reset
	docker buildx create --use --node wavefront_collector_builder
ifeq ($(RELEASE_TYPE), release)
	docker buildx build --platform linux/amd64,linux/arm64 --push \
	--build-arg BINARY_NAME=$(BINARY_NAME) --build-arg LDFLAGS="$(LDFLAGS)" \
	--pull -t $(PREFIX)/$(DOCKER_IMAGE):$(RELEASE_VERSION) -t $(PREFIX)/$(DOCKER_IMAGE):latest .
else
	docker buildx build --platform linux/amd64,linux/arm64 --push \
	--build-arg BINARY_NAME=$(BINARY_NAME) --build-arg LDFLAGS="$(LDFLAGS)" \
	--pull -t $(PREFIX)/$(DOCKER_IMAGE):$(RELEASE_VERSION)-rc-$(RC_NUMBER) .
endif

test-proxy-container: $(SEMVER_CLI_BIN)
	docker build \
	--build-arg BINARY_NAME=test-proxy --build-arg LDFLAGS="$(LDFLAGS)" \
	--pull -f $(REPO_DIR)/Dockerfile.test-proxy \
	-t $(PREFIX)/test-proxy:$(VERSION) .

test-proxy: peg $(REPO_DIR)/cmd/test-proxy/metric_grammar.peg.go clean fmt vet
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(ARCH)/test-proxy ./cmd/test-proxy/...

peg:
	@which peg > /dev/null || \
		(cd $(REPO_DIR)/..; GOARCH=$(ARCH) CGO_ENABLED=0 go get -u github.com/pointlander/peg)

$(GO_IMPORTS_BIN):
	@(cd $(REPO_DIR)/..; CGO_ENABLED=0 go get -u golang.org/x/tools/cmd/goimports)

semver-cli: $(SEMVER_CLI_BIN)

$(SEMVER_CLI_BIN):
	@(cd $(REPO_DIR)/..; CGO_ENABLED=0 go get -u github.com/davidrjonas/semver-cli)

%.peg.go: %.peg
	peg -switch -inline $<

#This rule need to be run on RHEL with podman installed.
container_rhel: build $(SEMVER_CLI_BIN)
	cp $(OUT_DIR)/$(ARCH)/$(BINARY_NAME) $(TEMP_DIR)
	cp LICENSE $(TEMP_DIR)/license.txt
	cp deploy/docker/Dockerfile-rhel $(TEMP_DIR)/Dockerfile
	cp deploy/examples/openshift-config.yaml $(TEMP_DIR)/collector.yaml
	podman build --pull -t $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(TEMP_DIR)
ifneq ($(OVERRIDE_IMAGE_NAME),)
	sudo docker tag $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(OVERRIDE_IMAGE_NAME)
endif

clean:
	@rm -f $(OUT_DIR)/$(ARCH)/$(BINARY_NAME)
	@rm -f $(OUT_DIR)/$(ARCH)/$(BINARY_NAME)-test
	@rm -f $(OUT_DIR)/$(ARCH)/test-proxy

token-check:
	@if [ -z ${WAVEFRONT_TOKEN} ]; then echo "Need to set WAVEFRONT_TOKEN" && exit 1; fi

proxy-test: token-check $(SEMVER_CLI_BIN)
ifeq ($(K8S_ENV), GKE)
	@(cd $(KUSTOMIZE_DIR) && ./test.sh nimba $(WAVEFRONT_TOKEN) $(VERSION) "us.gcr.io\/$(GCP_PROJECT)")
else ifeq ($(K8S_ENV), EKS)
	@(cd $(KUSTOMIZE_DIR) && ./test.sh nimba $(WAVEFRONT_TOKEN) $(VERSION) "$(ECR_ENDPOINT)\/tobs\/k8s\/saas")
else
	@(cd $(KUSTOMIZE_DIR) && ./test.sh nimba $(WAVEFRONT_TOKEN) $(VERSION))
endif

#Testing deployment and configuration changes, no code changes
deploy-test: token-check k8s-env clean-deployment deploy-targets push-images proxy-test

#Testing code, configuration and deployment changes
integration-test: token-check k8s-env clean-deployment deploy-targets containers delete-images push-images proxy-test

.PHONY: all fmt container clean release semver-cli
