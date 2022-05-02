#!/usr/bin/env bash

if [ $# -lt 1 ]; then
    echo "Usage: sort-json-keys-inplace.sh <directory glob or file with dashboard json>"
    exit 1
fi

sortKeysInPlace() {
    jq --sort-keys '.' $1 > $1.sorted
    mv $1.sorted $1
}
export -f sortKeysInPlace

target_dashboards=$@
find $target_dashboards -type f -exec bash -c 'sortKeysInPlace {}' \;
