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
    local borked_branch=$(git rev-parse --abbrev-ref HEAD)
    if [[ $borked_branch != "BORKWIP/"* ]]; then
        git checkout -b "BORKWIP/${borked_branch}"
        borked_branch="BORKWIP/${borked_branch}"
    fi

    git commit -m "BORKWIP: ${message}"
    git push --set-upstream origin "${borked_branch}"
}

function unBork() {
    local message=$1
    local borked_branch=$(git rev-parse --abbrev-ref HEAD)
    if [[ $borked_branch != "BORKWIP/"* ]]; then
        echo 'you are already unborked.'
        exit 1
    fi
    local unborked_branch=${borked_branch#BORKWIP/}

    gtc "UNBORK! : ${message}" || return 1
    echo "Stashing any uncommitted changes before checkout and rebase..."
    git stash
    git checkout "${unborked_branch}" || return 1
    git rebase -i "${borked_branch}" || return 1
    git push --set-upstream origin "${unborked_branch}" || return 1
    git branch -D ${borked_branch}
}
