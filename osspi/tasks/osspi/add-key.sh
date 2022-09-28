#!/usr/bin/env bash

function main {
  if [ ! "${GITHUB_KEY+defined}" = defined ]; then
    echo "Error: GITHUB_KEY undefined" >&2
    print_usage
    return 1
  elif [ -z "$GITHUB_KEY" ]; then
    echo "Error: GITHUB_KEY is blank" >&2
    return 1
  fi

  echo "$GITHUB_KEY" > git-key.pem
  chmod 600 git-key.pem

  eval "$(ssh-agent -s)"
  ssh-add git-key.pem

  if [ ! -d ~/.ssh ]; then
    mkdir -p ~/.ssh
  fi;

  mkdir -p ~/.ssh

  # add GitHub host key to known_hosts
  echo "github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==" >> ~/.ssh/known_hosts

  # add Bitbucket host key to known_hosts
  echo "bitbucket.org ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw==" >> ~/.ssh/known_hosts

  # Add VMware gitlab to known hosts
  ssh-keyscan -H gitlab.eng.vmware.com >> ~/.ssh/known_hosts

  # for pkg-config/pkg-config blob (in service-backup-release)
  ssh-keyscan -H gitlab.freedesktop.org >> ~/.ssh/known_hosts
}

function print_usage {
  echo "Usage: GITHUB_KEY=<github ssh key> add-key.sh"
}

main