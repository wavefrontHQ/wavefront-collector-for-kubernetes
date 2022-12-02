function green() {
  echo -e $'\e[32m'$1$'\e[0m'
}

function red() {
  echo -e $'\e[31m'$1$'\e[0m'
}

function print_msg_and_exit() {
  red "$1"
  exit 1
}

function pushd_check() {
  local d=$1
  pushd ${d} || print_msg_and_exit "Entering directory '${d}' with 'pushd' failed!"
}

function popd_check() {
  local d=$1
  popd || print_msg_and_exit "Leaving '${d}' with 'popd' failed!"
}

function wait_for_cluster_ready() {
  echo "Waiting for all Pods to be 'Ready'"
  while ! kubectl wait --for=condition=Ready pod --all -l exclude-me!=true --all-namespaces --timeout=5s  &> /dev/null; do
    printf "."
    sleep 1
  done
  echo " done."
}

function wait_for_namespace_created() {
	local namespace=$1
  printf "Waiting for namespace \"$1\" to be created ..."
  while ! kubectl create namespace collector-targets &> /dev/null; do
    printf "."
    sleep 1
  done
  echo " done."
}

function wait_for_namespaced_resource_created() {
    local namespace=$1
    local resource=$2
    while ! kubectl get --namespace $namespace $resource > /dev/null; do
        echo "Waiting for $resource in $namespace to be created"
        sleep 1
    done
}

function wait_for_cluster_resource_deleted() {
    local resource=$1
    while kubectl get $resource &> /dev/null; do
        echo "Waiting for $resource to be deleted"
        sleep 1
    done
}
