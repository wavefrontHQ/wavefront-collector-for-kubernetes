ECR_REPO_PREFIX=tobs/k8s/saas
WAVEFRONT_DEV_AWS_ACC_ID=095415062695
AWS_PROFILE=wavefront-dev
AWS_REGION=us-west-2
ECR_ENDPOINT=$(WAVEFRONT_DEV_AWS_ACC_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

ecr-host:
	echo $(ECR_ENDPOINT)/$(ECR_REPO_PREFIX)/wavefront-kubernetes-collector
target-eks:
	@aws eks --region $(AWS_REGION) update-kubeconfig --name k8s-saas-team-dev --profile $(AWS_PROFILE)
	@aws ecr get-login-password --region $(AWS_REGION) | sudo docker login --username AWS --password-stdin $(ECR_ENDPOINT)

push-to-ecr:
	make release PREFIX=$(ECR_ENDPOINT) DOCKER_IMAGE=$(ECR_REPO_PREFIX)/wavefront-kubernetes-collector
	docker tag $(PREFIX)/test-proxy:$(VERSION) $(ECR_ENDPOINT)/$(ECR_REPO_PREFIX)/test-proxy:$(VERSION)
    docker push $(ECR_ENDPOINT)/$(ECR_REPO_PREFIX)/test-proxy:$(VERSION)


delete-images-ecr:
	@aws ecr batch-delete-image --repository-name  $(ECR_REPO_PREFIX)/$(DOCKER_IMAGE) --image-ids imageTag=$(VERSION)