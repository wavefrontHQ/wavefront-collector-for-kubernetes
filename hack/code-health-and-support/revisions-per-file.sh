#!/usr/bin/env bash

REPO_ROOT=$(git rev-parse --show-toplevel)
OUTPUT_FILE=$REPO_ROOT/hack/code-health/commits-per-file.txt

go_files=$(find . -type f -name '*.go' -not -path '*/vendor/*')

echo '' > $OUTPUT_FILE
for go_file in $go_files ; do
    # count commits on file
    num_commits=$(git log -- $go_file | grep '^commit.*$' | wc -l | tr -d ' ')
    echo "$num_commits $go_file" >> $OUTPUT_FILE
done

sort --numeric-sort --reverse $OUTPUT_FILE --output=$OUTPUT_FILE
