#!/bin/bash

while true; do
    echo 'Starting collector in coverage mode'

    ./wavefront-collector.test -test.coverprofile=cover.out "$@" || exit 1
    
    echo "Collector restarting..."
done
