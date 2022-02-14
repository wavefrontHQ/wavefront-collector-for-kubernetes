
nuke-kind:
	kind delete cluster
	kind create cluster

nuke-kind-ha:
	kind delete cluster
	kind create cluster --config "make/k8s-envs/kind-ha.yml"

nuke-kind-with-advertise:
	kind delete cluster
	kind create cluster --config "make/k8s-envs/kind-with-advertise.yml"

kind-connect-to-cluster:
	kubectl config use kind-kind

push-to-kind:
	echo $(PREFIX)/$(DOCKER_IMAGE):$(VERSION)

	@kind load docker-image $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) --name kind

delete-images-kind:
	@docker exec -it kind-control-plane crictl rmi $(PREFIX)/$(DOCKER_IMAGE):$(VERSION) || true

target-kind:
	kubectl config use kind-kind