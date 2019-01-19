# Changes

The wavefront-kubernetes-collector leverages code from [heapster](https://github.com/kubernetes/heapster) for scraping metrics from the Kubelet's summary API.

Code adopted from heapster have been repackaged under the `internal` and `plugin` directories and retain their original copyright notices.

Additionally, the wavefront-kubernetes-collector includes the following enhancements:
1. Support for multiple sources
2. A prometheus source plugin for scraping prometheus metrics endpoints
3. Enhancements to the `Wavefront` sink plugin to support [direct ingestion](https://docs.wavefront.com/direct_ingestion.html)
4. A framework for [auto discovering](https://github.com/wavefrontHQ/wavefront-kubernetes-collector/blob/master/docs/discovery.md) pods and services that expose prometheus scrape targets.
