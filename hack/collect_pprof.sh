#! /bin/bash

# This script collects heap profiles from a pprof endpoint once a minute
# The collected heap profiles can be compared using "go tool pprof -base <source> <next>" to identify memory leaks etc
# It writes the profiles into a directory named "profiles-<starttime>"

profile_dir=profiles-`date +%s`
endpoint=localhost:9090

echo "collecting pprof heap profiles from ${endpoint}"
mkdir -p ${profile_dir}

x=0
while true; do
    curl localhost:9090/debug/pprof/heap > ${profile_dir}/heap.${x}.pprof
    let x=x+1
    sleep 60
done
