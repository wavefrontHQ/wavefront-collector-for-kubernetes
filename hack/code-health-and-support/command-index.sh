REPO_ROOT=$(git rev-parse --show-toplevel)

for f in $REPO_ROOT/hack/code-health-and-support/command/*.sh; do
    echo "sourcing '${f}'"
    source "${f}"
done
