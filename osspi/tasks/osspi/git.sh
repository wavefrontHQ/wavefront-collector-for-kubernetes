# shellcheck shell=bash

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "$here/debug.sh"

# is_repo_root
# return 0 is the directory is return root
# $1 - repo dir
function is_repo_root {
  declare -a git_command
  git_command=("git" "-C" "$1" "rev-parse" "--show-cdup")

  if is_debug_mode; then echo "Git command: ${git_command[*]}"; fi

  "${git_command[@]}"
}

# git_repo_name
# Get the name of a repo
# $1 - repo dir
# $2 - out var for repo name
function git_repo_name {
  declare -a git_command
  git_command=("git" "--git-dir=$1/.git" "remote" "get-url" "origin")

  local url
  url=$("${git_command[@]}")

  if is_debug_mode; then echo "Git command: ${git_command[*]}"; fi

  declare -a name_command
  name_command=("basename" "-s" ".git" "$url")
  if is_debug_mode; then echo "name command: ${name_command[*]}"; fi

  printf -v "$2" "$("${name_command[@]}")"
}

# git_repo_commit
# Get the commit of a repo
# $1 - repo dir
# $2 - out var for commit sha
function git_repo_commit {
  declare -a git_command
  git_command=("git" "-C" "$1" "rev-parse" "HEAD")

  if is_debug_mode; then echo "Git command: ${git_command[*]}"; fi

  printf -v "$2" "$("${git_command[@]}")"
}

# git_clone
# Git clone repo into directory
# $1 - ref
# $2 - repo
# $3 - shallow clone
# $4 - parent dir
# $5 - Out var for destination dir
function git_clone {
  local dst_dir
  dst_dir="$4/repo"
  mkdir -p "$dst_dir"

  declare -a command
  command+=("git" "clone")

  if [ ! "$3" = 'false' ]; then
    command+=("--depth=1")
  fi
  
  command+=("--single-branch" "--branch=$1" "$2" "$dst_dir")

  echo "Git clone command: ${command[*]}"

  local clone_log_file
  clone_log_file=$(mktemp /tmp/clone_log.XXXXXX.txt)
  echo "Clone logs will be saved to: $clone_log_file"

  if is_debug_mode; then
    if ! "${command[@]}" 2>&1 | tee "$clone_log_file"; then
      echo "Error: Clone failed, see logs: $clone_log_file" >&2
      return 1
    fi
  else
    if ! "${command[@]}" &> "$clone_log_file"; then
      echo "Error: Clone failed, see logs: $clone_log_file" >&2
      return 1
    fi
  fi

  if is_debug_mode; then echo "Cloned into: '$dst_dir'" >&2; fi

  printf -v "$5" "$dst_dir"
}
