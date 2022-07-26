function do_request() {
    local name=$1
    local url=$2
    mkdir -p /tmp/high-cardinality
    curl "$url" \
        -H 'Accept: text/event-stream' \
        -H "Authorization: Bearer $CLOUDHEALH_WAVEFRONT_TOKEN" \
        -w "$name [%{http_code}]: %{time_total}s\n" \
        -s -o "/tmp/high-cardinality/$name.dat"
    local stats=$(grep -A1 "event: stats" "/tmp/high-cardinality/$name.dat" | tail -n 1 | awk '{print $2}' | jq -r '"cardinality=" + (.stats.keys | tostring) + ", scanned=" + (.stats.points + .stats.distributions | tostring)')
    echo "$name: $stats"
}

WAVEFRONT_CLUSTER="cloudhealth"
K8S_CLUSTER="virginia-1.prod.cloudhealthtech.com"
START_TIME=$(date -v '-1d' '+%s')
END_TIME=$(date '+%s')

# Scenario 1, Original
do_request 'scenario1_original' "https://$WAVEFRONT_CLUSTER.wavefront.com/chart/streaming/v2?request=%7B%22id%22%3A1657746257740%2C%22queries%22%3A%5B%7B%22queryType%22%3A%22WQL%22%2C%22query%22%3A%22cumulativeHisto(sum(rate(ts(istio.request.duration.milliseconds.bucket%2C%20cluster%3D%5C%22%24%7Bcluster%7D%5C%22%20and%20destination_workload_namespace%3D%5C%22api-proxy%5C%22%20and%20destination_workload%3D%5C%22api-proxy*%5C%22%20%20and%20reporter%3D%5C%22destination%5C%22%20and%20response_code%3D%5C%222*%5C%22))%2C%20le))%22%2C%22name%22%3A%22Latency%22%2C%22secondaryAxis%22%3Afalse%7D%5D%2C%22includeObsoleteMetrics%22%3Afalse%2C%22expectedDataSpacing%22%3A60%2C%22queryParameters%22%3A%7B%22cluster%22%3A%22$K8S_CLUSTER%22%2C%22%24_globalFilter_%22%3A%22%22%7D%2C%22applyGlobalFilter%22%3Afalse%2C%22isLog%22%3Atrue%2C%22autoEvents%22%3Afalse%2C%22compareOffset%22%3A0%2C%22start%22%3A$START_TIME%2C%22end%22%3A$END_TIME%2C%22points%22%3A30%2C%22merging%22%3Atrue%2C%22limit%22%3A0%2C%22perSeriesStats%22%3Afalse%2C%22perSeriesRawStats%22%3Afalse%2C%22view%22%3A%22HISTOGRAM%22%7D&queryContext=%2Fcharts"

# Scenario 1, Aggregation
do_request 'scenario1_aggregation' "https://$WAVEFRONT_CLUSTER.wavefront.com/chart/streaming/v2?request=%7B%22id%22%3A1657746389966%2C%22queries%22%3A%5B%7B%22queryType%22%3A%22WQL%22%2C%22query%22%3A%22merge(hs(highcardtest.aggregation.istio.request.duration.milliseconds.m%2C%20cluster%3D%5C%22%24%7Bcluster%7D%5C%22%20and%20destination_workload_namespace%3D%5C%22api-proxy%5C%22%20and%20destination_workload%3D%5C%22api-proxy*%5C%22%20and%20reporter%3D%5C%22destination%5C%22%20and%20response_code%3D%5C%222*%5C%22))%22%2C%22name%22%3A%22Latency%22%2C%22secondaryAxis%22%3Afalse%7D%5D%2C%22includeObsoleteMetrics%22%3Afalse%2C%22expectedDataSpacing%22%3A60%2C%22queryParameters%22%3A%7B%22cluster%22%3A%22$K8S_CLUSTER%22%2C%22%24_globalFilter_%22%3A%22%22%7D%2C%22applyGlobalFilter%22%3Afalse%2C%22isLog%22%3Atrue%2C%22autoEvents%22%3Afalse%2C%22compareOffset%22%3A0%2C%22start%22%3A$START_TIME%2C%22end%22%3A$END_TIME%2C%22points%22%3A30%2C%22merging%22%3Atrue%2C%22limit%22%3A0%2C%22perSeriesStats%22%3Afalse%2C%22perSeriesRawStats%22%3Afalse%2C%22view%22%3A%22HISTOGRAM%22%7D&queryContext=%2Fcharts"

# Scenario 1, Downsample
do_request 'scenario1_downsample' "https://$WAVEFRONT_CLUSTER.wavefront.com/chart/streaming/v2?request=%7B%22id%22%3A1657746412056%2C%22queries%22%3A%5B%7B%22queryType%22%3A%22WQL%22%2C%22query%22%3A%22cumulativeHisto(sum(rate(ts(highcardtest.downsample.istio.request.duration.milliseconds.bucket%2C%20cluster%3D%5C%22%24%7Bcluster%7D%5C%22%20and%20destination_workload_namespace%3D%5C%22api-proxy%5C%22%20and%20destination_workload%3D%5C%22api-proxy*%5C%22%20and%20reporter%3D%5C%22destination%5C%22%20and%20response_code%3D%5C%222*%5C%22))%2C%20le))%22%2C%22name%22%3A%22Latency%22%2C%22secondaryAxis%22%3Afalse%7D%5D%2C%22includeObsoleteMetrics%22%3Afalse%2C%22expectedDataSpacing%22%3A60%2C%22queryParameters%22%3A%7B%22cluster%22%3A%22$K8S_CLUSTER%22%2C%22%24_globalFilter_%22%3A%22%22%7D%2C%22applyGlobalFilter%22%3Afalse%2C%22isLog%22%3Atrue%2C%22autoEvents%22%3Afalse%2C%22compareOffset%22%3A0%2C%22start%22%3A$START_TIME%2C%22end%22%3A$END_TIME%2C%22points%22%3A30%2C%22merging%22%3Atrue%2C%22limit%22%3A0%2C%22perSeriesStats%22%3Afalse%2C%22perSeriesRawStats%22%3Afalse%2C%22view%22%3A%22HISTOGRAM%22%7D&queryContext=%2Fcharts"

# Scenario 1, Better Source Tagging
do_request 'scenario1_source' "https://$WAVEFRONT_CLUSTER.wavefront.com/chart/streaming/v2?request=%7B%22id%22%3A1657746435832%2C%22queries%22%3A%5B%7B%22queryType%22%3A%22WQL%22%2C%22query%22%3A%22cumulativeHisto(sum(rate(ts(highcardtest.source.istio.request.duration.milliseconds.bucket%2C%20source%3D%5C%22%24%7Bcluster%7D%2Fapi-proxy%5C%22%20and%20destination_workload%3D%5C%22api-proxy*%5C%22%20and%20reporter%3D%5C%22destination%5C%22%20and%20response_code%3D%5C%222*%5C%22))%2C%20le))%22%2C%22name%22%3A%22Latency%22%2C%22secondaryAxis%22%3Afalse%7D%5D%2C%22includeObsoleteMetrics%22%3Afalse%2C%22expectedDataSpacing%22%3A60%2C%22queryParameters%22%3A%7B%22cluster%22%3A%22$K8S_CLUSTER%22%2C%22%24_globalFilter_%22%3A%22%22%7D%2C%22applyGlobalFilter%22%3Afalse%2C%22isLog%22%3Atrue%2C%22autoEvents%22%3Afalse%2C%22compareOffset%22%3A0%2C%22start%22%3A$START_TIME%2C%22end%22%3A$END_TIME%2C%22points%22%3A30%2C%22merging%22%3Atrue%2C%22limit%22%3A0%2C%22perSeriesStats%22%3Afalse%2C%22perSeriesRawStats%22%3Afalse%2C%22view%22%3A%22HISTOGRAM%22%7D&queryContext=%2Fcharts"

# Scenario 2, Original
do_request 'scenario2_original' "https://$WAVEFRONT_CLUSTER.wavefront.com/chart/streaming/v2?request=%7B%22id%22%3A1657746500035%2C%22queries%22%3A%5B%7B%22queryType%22%3A%22WQL%22%2C%22query%22%3A%22ts(%5C%22kubernetes.pod_container.restart_count%5C%22%2C%20cluster%3D%5C%22%24%7Bcluster%7D%5C%22)%22%2C%22name%22%3A%22A%22%2C%22secondaryAxis%22%3Afalse%7D%5D%2C%22summarizationStrategy%22%3A%22MEAN%22%2C%22includeObsoleteMetrics%22%3Afalse%2C%22expectedDataSpacing%22%3A60%2C%22queryParameters%22%3A%7B%22cluster%22%3A%22$K8S_CLUSTER%22%2C%22%24_globalFilter_%22%3A%22%22%7D%2C%22applyGlobalFilter%22%3Afalse%2C%22isLog%22%3Afalse%2C%22autoEvents%22%3Atrue%2C%22compareOffset%22%3A0%2C%22start%22%3A$START_TIME%2C%22end%22%3A$END_TIME%2C%22points%22%3A749%2C%22merging%22%3Atrue%2C%22limit%22%3A0%2C%22perSeriesStats%22%3Afalse%2C%22perSeriesRawStats%22%3Afalse%7D&queryContext=%2Fcharts"

# Scenario 2, Downsample
do_request 'scenario2_downsample' "https://$WAVEFRONT_CLUSTER.wavefront.com/chart/streaming/v2?request=%7B%22id%22%3A1657746502806%2C%22queries%22%3A%5B%7B%22queryType%22%3A%22WQL%22%2C%22query%22%3A%22ts(%5C%22highcardtest.downsample.kubernetes.pod_container.restart_count%5C%22%2C%20cluster%3D%5C%22%24%7Bcluster%7D%5C%22)%22%2C%22name%22%3A%22A%22%2C%22secondaryAxis%22%3Afalse%7D%5D%2C%22summarizationStrategy%22%3A%22MEAN%22%2C%22includeObsoleteMetrics%22%3Afalse%2C%22expectedDataSpacing%22%3A60%2C%22queryParameters%22%3A%7B%22cluster%22%3A%22$K8S_CLUSTER%22%2C%22%24_globalFilter_%22%3A%22%22%7D%2C%22applyGlobalFilter%22%3Afalse%2C%22isLog%22%3Afalse%2C%22autoEvents%22%3Atrue%2C%22compareOffset%22%3A0%2C%22start%22%3A$START_TIME%2C%22end%22%3A$END_TIME%2C%22points%22%3A749%2C%22merging%22%3Atrue%2C%22limit%22%3A0%2C%22perSeriesStats%22%3Afalse%2C%22perSeriesRawStats%22%3Afalse%7D&queryContext=%2Fcharts"

# Scenario 2, Better Source Tagging
do_request 'scenario2_source' "https://$WAVEFRONT_CLUSTER.wavefront.com/chart/streaming/v2?request=%7B%22id%22%3A1657746503763%2C%22queries%22%3A%5B%7B%22queryType%22%3A%22WQL%22%2C%22query%22%3A%22ts(%5C%22highcardtest.source.kubernetes.pod_container.restart_count%5C%22%2C%20source%3D%5C%22%24%7Bcluster%7D%5C%22)%22%2C%22name%22%3A%22A%22%2C%22secondaryAxis%22%3Afalse%7D%5D%2C%22summarizationStrategy%22%3A%22MEAN%22%2C%22includeObsoleteMetrics%22%3Afalse%2C%22expectedDataSpacing%22%3A60%2C%22queryParameters%22%3A%7B%22cluster%22%3A%22$K8S_CLUSTER%22%2C%22%24_globalFilter_%22%3A%22%22%7D%2C%22applyGlobalFilter%22%3Afalse%2C%22isLog%22%3Afalse%2C%22autoEvents%22%3Atrue%2C%22compareOffset%22%3A0%2C%22start%22%3A$START_TIME%2C%22end%22%3A$END_TIME%2C%22points%22%3A749%2C%22merging%22%3Atrue%2C%22limit%22%3A0%2C%22perSeriesStats%22%3Afalse%2C%22perSeriesRawStats%22%3Afalse%7D&queryContext=%2Fcharts"