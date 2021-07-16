ECR_REPO_PREFIX=tobs/k8s/saas
WAVEFRONT_DEV_AWS_ACC_ID=095415062695
AWS_PROFILE=wavefront-dev
AWS_REGION=us-west-2
ECR_ENDPOINT=${WAVEFRONT_DEV_AWS_ACC_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

target-eks:
	@aws eks --region $(AWS_REGION) update-kubeconfig --name k8s-saas-team-dev --profile $(AWS_PROFILE)
	@aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(ECR_ENDPOINT)

push-to-ecr:
	@./hack/make/push-container.sh $(PREFIX) $(VERSION) $(ECR_ENDPOINT) $(ECR_REPO_PREFIX)

delete-images-ecr:
	@aws ecr batch-delete-image --repository-name  $(ECR_REPO_PREFIX)/$(DOCKER_IMAGE) --image-ids imageTag=$(VERSION)