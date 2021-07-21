ECR_REPO_PREFIX=tobs/k8s/saas
WAVEFRONT_DEV_AWS_ACC_ID=095415062695
AWS_PROFILE=wavefront-dev
AWS_REGION=us-west-2
ECR_ENDPOINT=$(WAVEFRONT_DEV_AWS_ACC_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
COLLECTOR_ECR_REPO=$(ECR_REPO_PREFIX)/wavefront-kubernetes-collector
TEST_PROXY_ECR_REPO=$(ECR_REPO_PREFIX)/test-proxy

ecr-host:
	echo $(ECR_ENDPOINT)/$(ECR_REPO_PREFIX)/wavefront-kubernetes-collector

docker-login-eks:
	@aws ecr get-login-password --region $(AWS_REGION) |  docker login --username AWS --password-stdin $(ECR_ENDPOINT)

target-eks: docker-login-eks
	@aws eks --region $(AWS_REGION) update-kubeconfig --name k8s-saas-team-dev --profile $(AWS_PROFILE)

push-to-ecr: docker-login-eks
	docker tag $(PREFIX)/test-proxy:$(VERSION) $(ECR_ENDPOINT)/$(TEST_PROXY_ECR_REPO):$(VERSION)
	docker push $(ECR_ENDPOINT)/$(ECR_REPO_PREFIX)/test-proxy:$(VERSION)

	@aws --no-cli-pager ecr describe-images --region $(AWS_REGION) --repository-name $(COLLECTOR_ECR_REPO) --image-ids imageTag=$(VERSION) > /dev/null;\
	EXIT_CODE=$$?;\
	if [ $$EXIT_CODE -ne 0 ]; then\
	    make release PREFIX=$(ECR_ENDPOINT) DOCKER_IMAGE=$(COLLECTOR_ECR_REPO);\
	fi

delete-images-ecr:
	aws --no-cli-pager ecr batch-delete-image --region $(AWS_REGION) --repository-name  $(COLLECTOR_ECR_REPO)  --image-ids imageTag=$(VERSION) imageTag=latest || true
	aws --no-cli-pager ecr batch-delete-image --region $(AWS_REGION) --repository-name  $(TEST_PROXY_ECR_REPO) --image-ids imageTag=$(VERSION) || true

	# removes untagged images for multi-platform/arch build
	aws ecr list-images --region $(AWS_REGION) --repository-name $(COLLECTOR_ECR_REPO) --filter "tagStatus=UNTAGGED" --output json > /tmp/ecr-delete-list.json || true
	aws --no-cli-pager ecr batch-delete-image --region  $(AWS_REGION) --repository-name $(COLLECTOR_ECR_REPO) --cli-input-json file:///tmp/ecr-delete-list.json || true
	rm /tmp/ecr-delete-list.json