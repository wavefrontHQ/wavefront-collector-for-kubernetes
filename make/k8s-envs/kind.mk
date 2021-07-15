
nuke-kind:
	kind delete cluster
	kind create cluster

kind-connect-to-cluster:
	kubectl config use kind-kind

push-to-kind:
	echo $(PREFIX)/$(DOCKER_IMAGE):$(VERSION)

	@kind load docker-image $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) --name kind
	@kind load docker-image $(PREFIX)/test-proxy:$(VERSION) --name kind

delete-images-kind:
	@docker exec -it kind-control-plane crictl rmi $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) || true
	@docker exec -it kind-control-plane crictl rmi $(PREFIX)/test-proxy:$(VERSION) || true