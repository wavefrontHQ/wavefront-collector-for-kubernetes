# Useful stuff used in Norsk to OSSPI debugging

PIPELINE=collector-osspi
TARGET=runway-ci-sfo

function dockerScanFromFailedBuild() {
    local task_name=$1
    local build_num=$2
    local build_step=$3
    fly -t $TARGET hijack -j $PIPELINE/$task_name -b $build_num -s $build_step tar zcvf docker_scan.tar.gz docker_scan
    fly -t $TARGET hijack -j $PIPELINE/$task_name -b $build_num -s $build_step cat docker_scan.tar.gz > docker_scan.tar.gz
}

alias fsp='fly -t runway-ci-sfo set-pipeline -p collector-osspi -c osspi/pipeline.yaml'
