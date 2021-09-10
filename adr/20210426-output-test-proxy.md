# 20210426 Output Test Proxy

## Context

The output test was introduced in order to provide an end-to-end integration test. The intent was to facilitate releasing larger architectural changes with confidence. The first implementation reported metrics and diff them against a "golden" copy. We logged all recorded metrics to the pod logs, scraped them, sliced them up, and then diffed them using bash tools. This presented a couple of problems

* not everyone on the team wanted to use bash for our diffing logic (go seemed like a natural alternative)
* we wanted to be able to expand this test to more diverse environments which would be more challenging without more post-processing to ensure that various environment specific tags were not included in the diff
* it required us to excercise test code instead of production code for the reporting logic

## Decision

We decided to develop a test proxy written in go that was a drop in replacement for the wavefront proxy. This test proxy recorded all metrics in memory and exposed an API to diff them.

## Status
[Implemented](https://github.com/wavefrontHQ/wavefront-collector-for-kubernetes/pull/143)

## Consequences

This allowed us to run the collector in a manner that exercised production code paths. In addition, we could write arbitrary diffing logic that was much more flexible than our previous bash scripts. We are currently porting the the output test to GKE.

The downside is we have to maintain a server that understands and stays in sync with the wavefront wire format. In addition, the test proxy is more code that we have to maintain. It also won't tell us if the wavefront proxy wire format ever changes.