#!/bin/bash -e

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
# helm
#
curl https://get.helm.sh/helm-v3.6.3-linux-amd64.tar.gz | tar xz --to-stdout linux-amd64/helm | sudo tee /usr/local/bin/helm >/dev/null
sudo chmod +x /usr/local/bin/helm

#
# kustomize
#
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
chmod +x ./kustomize
sudo mv ./kustomize /usr/local/bin

#
# semver cli
#
git config --global http.sslVerify false
make semver-cli
git config --global http.sslVerify true

#
# jq
#
if ! [ -x "$(command -v jq)" ]; then
  mkdir -p "$HOME/bin"
  curl -L https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 --output "$HOME/bin/jq"
  chmod +x "$HOME/bin/jq"
fi