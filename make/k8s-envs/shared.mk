deploy-targets:
	@(cd $(DEPLOY_DIR) && ./deploy-targets.sh)

clean-targets:
	@(cd $(DEPLOY_DIR) && ./uninstall-targets.sh)

k9s:
	watch -n 1 k9s

clean-deployment:
	@(cd $(DEPLOY_DIR) && ./uninstall-wavefront-helm-release.sh)
	@(cd $(TEST_DIR) && ./clean-deploy.sh)

k8s-env:
	@echo "\033[92mK8s Environment: $(shell kubectl config current-context)\033[0m"

k8s-nodes-arch:
	kubectl get nodes --label-columns='kubernetes.io/arch'

clean-cluster: clean-targets clean-deployment

push-images:
ifeq ($(K8S_ENV), Kind)
	make push-to-kind
else
	make publish
endif

cover-push-images:
ifeq ($(K8S_ENV), Kind)
	make cover-push-to-kind
else
	make publish
endif

delete-images:
ifeq ($(K8S_ENV), Kind)
	make delete-images-kind
endif