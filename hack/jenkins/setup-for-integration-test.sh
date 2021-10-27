#!/bin/bash -ex

#
# gcloud
#
if ! [ -x "$(command -v gcloud)" ]; then
  curl https://sdk.cloud.google.com > install.sh
  chmod +x ./install.sh
  sudo PREFIX=$HOME ./install.sh --disable-prompts >/dev/null;
fi

gcloud auth activate-service-account --key-file "$GCP_CREDS"
gcloud config set project wavefront-gcp-dev

#
# docker-credential-gcr
#
curl -fsSL "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v2.0.0/docker-credential-gcr_linux_amd64-2.0.0.tar.gz" \
  | tar xz --to-stdout ./docker-credential-gcr | sudo tee /usr/local/bin/docker-credential-gcr >/dev/null
sudo chmod +x /usr/local/bin/docker-credential-gcr
docker-credential-gcr config --token-source="gcloud"
docker-credential-gcr configure-docker --registries="us.gcr.io"
(echo "https://us.gcr.io" | docker-credential-gcr get >/dev/null) \
  || (echo "docker credentials not configured properly"; exit 1)

#
# kubectl
#
curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl

#
# jq
#
if ! [ -x "$(command -v jq)" ]; then
  curl -H "Authorization: token ${GITHUB_CREDS_PSW}" -L "https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64" > ./jq
  chmod +x ./jq
  sudo mv ./jq /usr/local/bin
fi

#
# helm
#
curl https://get.helm.sh/helm-v3.6.3-linux-amd64.tar.gz | tar xz --to-stdout linux-amd64/helm | sudo tee /usr/local/bin/helm >/dev/null
sudo chmod +x /usr/local/bin/helm

#
# kustomize
#
if ! [ -x "$(command -v kustomize)" ]; then
  curl -H "Authorization: token ${GITHUB_CREDS_PSW}" -L -s "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.4.0/kustomize_v4.4.0_linux_amd64.tar.gz" \
    | tar xz --to-stdout ./usr/local/bin/kustomize \
    | sudo tee /usr/local/bin/kustomize >/dev/null
  sudo chmod +x /usr/local/bin/kustomize
fi

#
# semver cli
#
git config --global http.sslVerify false
make semver-cli
git config --global http.sslVerify true
