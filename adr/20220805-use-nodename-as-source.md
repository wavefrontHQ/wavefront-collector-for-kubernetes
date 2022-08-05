# 20220805 Use Node Name as source value for Kubernetes metrics

## Context
In Wavefront source is a special point tag used to identify a unique application, host, container, or instance that emits metrics. It is also special in the way Wavefront handles queries based on the source value. So optimizing the source value is an important step in improving query performance at scale. 

The team experimented with the below source values and measure query performance at scale by using cloudhealth's environment.
* source=cluster

This had the worst performance, this is due to the fact that the query engine parallelizes queries across different source combinations, so using distinct sources can decrease query latency.
* source=cluster+namespace

This improved performance compared to previous but worse compared to the next one.
* source=nodename

This had the best performance. Compared to source=cluster+namespace, we attribute the improved query performance to the fact that workloads were distributed equally across nodes, so the number of distinct timeseries were identical for a given metric + source. Theoritically, this meant parallel scans for each source will have identical no. of points to be scanned and hence nodename would be an ideal source value for optimizing query performance at scale.

## Decision
Given how Wavefront backend uses source for querying, we concluded that source=nodename is a good generic choice for source value. However, depending on the data shape and customer query usage an optimization to source value might be possible.   


