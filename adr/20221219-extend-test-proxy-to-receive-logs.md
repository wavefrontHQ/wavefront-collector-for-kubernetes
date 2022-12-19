# 20221219 Extend Test Proxy to Receive Logs

## Context
The team needed a way to write logging integration tests.

The team considered, extending the existing test proxy or creating a new one for logging only.

## Decision
* Extend the existing test proxy to receive logs and API to assert they were in the expected format.
* A run mode: logs vs. metrics was added so that we could use the same port similar to the real proxy. 
* In addition, we move code from cmd directory into package files to make it better factored and easier to test. We confirmed the test files will not be included in the wavefront binary.

## Status
* [Implemented](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/pull/549)

## Future concerns
* Duplication of the test proxy yaml and start scripts between this repo and the [Wavefront K8s Operator](https://github.com/wavefrontHQ/wavefront-operator-for-kubernetes) repo. We should consider consolidating 
them as part of the mono repo effort. 
