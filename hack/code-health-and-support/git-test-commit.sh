# NOTE this file should be sourced

function gtc() {
    local message=$1
    # TODO automatically launch kind if not running
    local kind_clusters=$(kind get clusters 2>&1 1>/dev/null)

    local start_time=`date +%s`

    if [ "${kind_clusters}" = "No kind clusters found." ]; then
        make nuke-kind
    fi

    make tests || return 1
    make integration-test || return 1

    local end_time=`date +%s`

    local total_runtime=$((end_time-start_time))
    git status
    git commit -m "${message}

    - total test runtime: ${total_runtime}s"
    git push
}

function borkWIP() {
    local message=$1

    local branch=$(git rev-parse --abbrev-ref HEAD)
    if [[ $branch != "BORKWIP/"* ]]; then
        git checkout -b "BORKWIP/${branch}"
        branch="BORKWIP/${branch}"
    fi

    git commit -m "BORKWIP: ${message}"
    git push --set-upstream origin "${branch}"
}
