#!/bin/bash

while true; do
    echo 'Starting collector in coverage mode'

    COLLECTOR_COVERAGE_ARGS="'$*'" go test ./... -cover -covermode=count -coverpkg=./... -coverprofile=cover.out || exit 1
    
    echo "Collector restarting..."
done
