deploy-targets:
	@(cd $(DEPLOY_DIR) && ./deploy-targets.sh)

clean-targets:
	@(cd $(DEPLOY_DIR) && ./uninstall-targets.sh)

k9s:
	watch -n 1 k9s

clean-deployment:
	@(cd $(DEPLOY_DIR) && ./uninstall-wavefront-helm-release.sh)
	@(cd $(KUSTOMIZE_DIR) && ./clean-deploy.sh)

k8s-env:
	@echo "\033[92mK8s Environment: $(shell kubectl config current-context)\033[0m"

k8s-nodes-arch:
	kubectl get nodes --label-columns='kubernetes.io/arch'

clean-cluster: clean-targets clean-deployment

push-images:
ifeq ($(K8S_ENV), GKE)
	make push-to-gcr
else ifeq ($(K8S_ENV), EKS)
	make push-to-ecr
else
	make push-to-kind
endif

push-test-proxy-image:
ifeq ($(K8S_ENV), GKE)
	make push-test-proxy-to-gcr
else
	echo "Not implemented"
	exit 1
endif

delete-images:
ifeq ($(K8S_ENV), GKE)
	make delete-images-gcr
else ifeq ($(K8S_ENV), EKS)
	make delete-images-ecr
else
	make delete-images-kind
endif
