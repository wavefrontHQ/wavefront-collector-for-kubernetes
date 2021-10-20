#!/bin/bash -e

Help() {
  # Display Help
  echo "Generate slack CI regression message."
  echo
  echo "Syntax: $0 [-u|n|b|h]"
  echo "options:"
  printf "\t-u     Build URL.\n"
  printf "\t-n     Build Number.\n"
  printf "\t-b     Git branch.\n"
  printf "\t-h     Print this help.\n"
  echo
}

main() {
  cd "$(dirname "$0")"
  local BUILD_URL=
  local BUILD_NUMBER=
  local GIT_BRANCH=
  while getopts ":hu:n:b:" option; do
    case $option in
    h) # display Help
      Help
      exit
      ;;
    u)
      BUILD_URL=$OPTARG
      ;;
    n)
      BUILD_NUMBER=$OPTARG
      ;;
    b)
      GIT_BRANCH=$OPTARG
      ;;
    \?) # Invalid option
      echo "Error: Invalid option -$OPTARG. Use -h to see valid options."
      exit 1
      ;;
    esac
  done

  if [ ! "$BUILD_URL" ] || [ ! "$BUILD_NUMBER" ] || [ ! "$GIT_BRANCH" ]; then
    echo "Need to specify all options: build url (-b), build number (-n) and git branch (-r). Use -h to see valid options."
    exit 1
  fi

  echo "Build <${BUILD_URL}|#${BUILD_NUMBER}> failed the build on ${GIT_BRANCH}!"
}

main "$@"
