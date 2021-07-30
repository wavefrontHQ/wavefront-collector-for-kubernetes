
nuke-kind:
	kind delete cluster
	kind create cluster

kind-connect-to-cluster:
	kubectl config use kind-kind

ensure-kind-cluster-running:
	@if [[ "$(shell kind get clusters 2>&1)" == "No kind clusters found." ]]; then \
		echo "You don't have a kind cluster yet! Running 'kind create cluster' for you so you can run the integration tests.";\
		kind create cluster;\
		echo "Run 'WAVEFRONT_API_KEY=<wavefront-api-key> make integration-test' from the '~/workspace/wavefront-collector-for-kubernetes' directory to test setup.";\
	else \
		echo 'Kind cluster is already running!';\
	fi

push-to-kind:
	echo $(PREFIX)/$(DOCKER_IMAGE):$(VERSION)

	@kind load docker-image $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) --name kind
	@kind load docker-image $(PREFIX)/test-proxy:$(VERSION) --name kind

delete-images-kind:
	@docker exec -it kind-control-plane crictl rmi $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) || true
	@docker exec -it kind-control-plane crictl rmi $(PREFIX)/test-proxy:$(VERSION) || true
