#!/usr/bin/env bash
source hack/make/_script-tools.sh

if [[ -z ${K8S_ENV} ]]; then
  print_msg_and_exit 'K8S_ENV required but was empty'
  #K8S_ENV=$DEFAULT_K8S_ENV
fi

if [[ -z ${GCP_PROJECT} ]]; then
  print_msg_and_exit 'GCP_PROJECT required but was empty'
  #GCP_PROJECT=$DEFAULT_GCP_PROJECT
fi

if [[ -z ${VERSION} ]]; then
  print_msg_and_exit 'VERSION required but was empty'
  #VERSION=$DEFAULT_VERSION
fi

if [[ -z ${PREFIX} ]]; then
  print_msg_and_exit 'PREFIX required but was empty'
  #PREFIX=$DEFAULT_PREFIX
fi

if [[ -z ${DOCKER_IMAGE} ]]; then
  print_msg_and_exit 'DOCKER_IMAGE required but was empty'
  #DOCKER_IMAGE=$DEFAULT_DOCKER_IMAGE
fi

if [[ -z ${VERSION} ]]; then
  print_msg_and_exit 'VERSION required but was empty'
  #VERSION=$DEFAULT_VERSION
fi

ECR_REPO_PREFIX=tobs/k8s/saas

# commands ...
if [ ${K8S_ENV} == "GKE" ]; then
  gcloud container images delete us.gcr.io/${GCP_PROJECT}/test-proxy:${VERSION} --quiet &>/dev/null || true
  gcloud container images delete us.gcr.io/${GCP_PROJECT}/wavefront-kubernetes-collector:${VERSION} --quiet &>/dev/null || true
elif [ ${K8S_ENV} == "EKS" ]; then
  aws ecr batch-delete-image --repository-name  ${ECR_REPO_PREFIX}/${DOCKER_IMAGE} --image-ids imageTag=${VERSION}
else
  docker exec -it kind-control-plane crictl rmi ${PREFIX}/${DOCKER_IMAGE}:${VERSION} || true
	kind load docker-image ${PREFIX}/${DOCKER_IMAGE}:${VERSION} --name kind

	docker exec -it kind-control-plane crictl rmi ${PREFIX}/test-proxy:${VERSION} || true
	kind load docker-image ${PREFIX}/test-proxy:${VERSION} --name kind
fi
