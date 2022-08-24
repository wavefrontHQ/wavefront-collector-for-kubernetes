# NOTE this file should be sourced

function gtc() {
    local message=$1
    # TODO automatically launch kind if not running
    local kind_clusters=$(kind get clusters 2>&1 1>/dev/null)

    if [ "${kind_clusters}" = "No kind clusters found." ]; then
        make nuke-kind
    fi
    make tests || return 1
    make integration-test || return 1

    git status
    git commit -m "${message}"
}
