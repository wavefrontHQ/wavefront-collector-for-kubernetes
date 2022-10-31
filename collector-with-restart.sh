#!/bin/bash

function sleepilyAssassinate() {
    sleep $(( 60*5 ))
    curl 'localhost:19999'
}

while true; do
    echo 'Starting collector in coverage mode'

    sleepilyAssassinate &
    ./wavefront-collector.test -test.coverprofile=cover.out "$@" || exit 1
    
    echo "Collector restarting..."
done

# TODO squash and merge
