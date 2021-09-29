#!/usr/bin/env bash
set -e
if [[ ! -f "$HOME/.docker/cli-plugins/docker-buildx" ]]; then
  echo "installing docker buildx..."
  wget -q -O docker-buildx https://github.com/docker/buildx/releases/download/v0.6.3/buildx-v0.6.3.linux-amd64
  chmod a+x docker-buildx
  [[ ! -d "$HOME/.docker/cli-plugins" ]] && mkdir -p ~/.docker/cli-plugins
  mv docker-buildx ~/.docker/cli-plugins
  echo "successfully installed docker buildx: $(docker buildx version)"
else
  echo "buildx already installed: $(docker buildx version)"
fi
