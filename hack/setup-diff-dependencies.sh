#!/bin/bash -e

if [[ ! -f "$HOME/.osspicli/osspi/osspi" ]]; then
  echo "installing osspi..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    bash -c "$(curl -fsSL https://build-artifactory.eng.vmware.com/osspicli-local/beta/osspicli-darwin/install.sh)"
  else
    bash -c "$(curl -fsSL https://build-artifactory.eng.vmware.com/osspicli-local/beta/osspicli/install.sh)"
  fi
  echo "successfully installed osspi: $($HOME/.osspicli/osspi/osspi --version)"
else
  echo "osspi already installed: $($HOME/.osspicli/osspi/osspi --version)"
fi

if ! [ -x "$(command -v jq)" ]; then
  curl -H "Authorization: token ${GITHUB_CREDS_PSW}" -L "https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64" > ./jq
  chmod +x ./jq
  sudo mv ./jq /usr/local/bin
fi