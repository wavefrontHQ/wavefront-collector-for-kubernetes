PREFIX?=projects.registry.vmware.com/tanzu_observability_keights_saas
DOCKER_IMAGE?=kubernetes-collector-snapshot
ARCH?=amd64
WAVEFRONT_CLUSTER?=nimba

REPO_DIR=$(shell git rev-parse --show-toplevel)
TEST_DIR=$(REPO_DIR)/hack/test
DEPLOY_DIR=$(REPO_DIR)/hack/test/deploy
OUT_DIR?=$(REPO_DIR)/_output
INTEGRATION_TEST_TYPE?=default

BINARY_NAME=wavefront-collector

RELEASE_TYPE?=dev
RC_NUMBER?=1
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GIT_HUB_REPO=wavefrontHQ/wavefront-collector-for-kubernetes

K8S_ENV=$(shell cd $(DEPLOY_DIR) && ./get-k8s-cluster-env.sh)

ifndef TEMP_DIR
TEMP_DIR:=$(shell mktemp -d /tmp/wavefront.XXXXXX)
endif

GO_IMPORTS_BIN:=$(if $(which goimports),$(which goimports),$(GOPATH)/bin/goimports)
SEMVER_CLI_BIN:=$(if $(which semver-cli),$(which semver-cli),$(GOPATH)/bin/semver-cli)

VERSION_POSTFIX?=-dev-$(shell whoami)-$(shell git rev-parse --short HEAD)
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

# TODO: exclude certain keys from sorting
# because we want them to be at the top and visible when we open the file!
sort-integrations-keys:
	# TODO: uncomment to run this on all of our dashboards when we're comfortable
	@#$(REPO_DIR)/hack/sort-json-keys-inplace.sh $(HOME)/workspace/integrations/kubernetes/dashboards/*.json
	@$(REPO_DIR)/hack/sort-json-keys-inplace.sh $(HOME)/workspace/integrations/kubernetes/dashboards/integration-kubernetes-control-plane.json

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

container: $(SEMVER_CLI_BIN)
	# Run build in a container in order to have reproducible builds
	docker build \
	-f $(REPO_DIR)/Dockerfile.non-cross-platform \
	--build-arg BINARY_NAME=$(BINARY_NAME) --build-arg LDFLAGS="$(LDFLAGS)" \
	--pull -t $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) .
ifneq ($(OVERRIDE_IMAGE_NAME),)
	docker tag $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(OVERRIDE_IMAGE_NAME)
endif

BUILDER_SUFFIX=$(shell echo $(PREFIX) | cut -d '/' -f1)

publish:
	docker buildx create --use --node wavefront_collector_builder_$(BUILDER_SUFFIX)
ifeq ($(RELEASE_TYPE), release)
	docker buildx build --platform linux/amd64,linux/arm64 --push \
	--build-arg BINARY_NAME=$(BINARY_NAME) --build-arg LDFLAGS="$(LDFLAGS)" \
	--pull -t $(PREFIX)/$(DOCKER_IMAGE):$(RELEASE_VERSION) -t $(PREFIX)/$(DOCKER_IMAGE):latest .
else ifeq ($(RELEASE_TYPE), rc)
	docker buildx build --platform linux/amd64,linux/arm64 --push \
	--build-arg BINARY_NAME=$(BINARY_NAME) --build-arg LDFLAGS="$(LDFLAGS)" \
	--pull -t $(PREFIX)/$(DOCKER_IMAGE):$(RELEASE_VERSION)-rc-$(RC_NUMBER) .
else
	docker buildx build --platform linux/amd64,linux/arm64 --push \
	--build-arg BINARY_NAME=$(BINARY_NAME) --build-arg LDFLAGS="$(LDFLAGS)" \
	--pull -t $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) .
endif

test-proxy-container: $(SEMVER_CLI_BIN)
	docker build \
	--build-arg BINARY_NAME=test-proxy --build-arg LDFLAGS="$(LDFLAGS)" \
	--pull -f $(REPO_DIR)/Dockerfile.test-proxy \
	-t $(PREFIX)/test-proxy:$(VERSION) -t $(PREFIX)/test-proxy:latest .

publish-test-proxy:  test-proxy-container
	docker push $(PREFIX)/test-proxy:latest
	docker push $(PREFIX)/test-proxy:$(VERSION)

test-proxy: peg $(REPO_DIR)/cmd/test-proxy/metric_grammar.peg.go clean fmt vet
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(ARCH)/test-proxy ./cmd/test-proxy/...

peg:
	@which peg > /dev/null || \
		(cd $(REPO_DIR)/..; GOARCH=$(ARCH) CGO_ENABLED=0 go install github.com/pointlander/peg@latest)

$(GO_IMPORTS_BIN):
	@(cd $(REPO_DIR)/..; CGO_ENABLED=0 go install golang.org/x/tools/cmd/goimports@latest)

semver-cli: $(SEMVER_CLI_BIN)

$(SEMVER_CLI_BIN):
	@(cd $(REPO_DIR)/..; CGO_ENABLED=0 go install github.com/davidrjonas/semver-cli@latest)

%.peg.go: %.peg
	peg -switch -inline $<

container_rhel: $(SEMVER_CLI_BIN)
	docker build \
		-f $(REPO_DIR)/deploy/docker/Dockerfile-rhel \
		--build-arg COLLECTOR_VERSION=$(RELEASE_VERSION) --build-arg LDFLAGS="$(LDFLAGS)" \
		-t $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) .
ifneq ($(OVERRIDE_IMAGE_NAME),)
	sudo docker tag $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(OVERRIDE_IMAGE_NAME)
endif

push_rhel_redhat_connect: container_rhel
	docker tag  $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) $(PREFIX)/$(DOCKER_IMAGE):$(RELEASE_VERSION)-rc
	docker push $(PREFIX)/$(DOCKER_IMAGE):$(RELEASE_VERSION)-rc

clean:
	@rm -f $(OUT_DIR)/$(ARCH)/$(BINARY_NAME)
	@rm -f $(OUT_DIR)/$(ARCH)/$(BINARY_NAME)-test
	@rm -f $(OUT_DIR)/$(ARCH)/test-proxy

token-check:
	@if [ -z ${WAVEFRONT_TOKEN} ]; then echo "Need to set WAVEFRONT_TOKEN" && exit 1; fi

proxy-test: token-check $(SEMVER_CLI_BIN)
	(cd $(TEST_DIR) && ./test-integration.sh $(WAVEFRONT_CLUSTER) $(WAVEFRONT_TOKEN) $(VERSION) $(INTEGRATION_TEST_TYPE))

#Testing deployment and configuration changes, no code changes
deploy-test: token-check k8s-env clean-deployment deploy-targets proxy-test

#Testing code, configuration and deployment changes
integration-test: token-check k8s-env clean-deployment deploy-targets delete-images push-images proxy-test

# creating this as separate and distinct for now,
# but would like to recombine as a flag on integration-test
integration-test-rhel: token-check k8s-env clean-deployment deploy-targets
	VERSION=$(VERSION)-rhel make container_rhel test-proxy-container delete-images push-images proxy-test

# create a new branch from main
# usage: make branch JIRA=XXXX OR make branch NAME=YYYY
branch:
	$(eval NAME := $(if $(JIRA),K8SAAS-$(JIRA),$(NAME)))
	@if [ -z "$(NAME)" ]; then \
		echo "usage: make branch JIRA=XXXX OR make branch NAME=YYYY"; \
		exit 1; \
	fi
	git stash
	git checkout main
	git pull
	git checkout -b $(NAME)

git-rebase:
	git fetch origin
	git rebase origin/main
	git log --oneline -n 10

clean-cluster:
	(cd $(TEST_DIR) && ./clean-cluster.sh)

# list the available makefile targets
list:
	@sh -c "$(MAKE) -p no_targets__ | awk -F':' '/^[a-zA-Z0-9][^\$$#\/\\t=]*:([^=]|$$)/ {split(\$$1,A,/ /);for(i in A)print A[i]}' | grep -v '__\$$' | sort"

.PHONY: all fmt container clean release semver-cli no_targets__ list
