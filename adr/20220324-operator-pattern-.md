# 20220324 Choosing Operator Pattern or Framework for unified install
## Context

- Inorder to make installation simpler and more consistent we decided to look towards having an operator to install and manage both our components(collector and proxy).
- Frameworks considered:
  - [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) 
  - [operator-sdk](https://github.com/operator-framework/operator-sdk)
  - [operator-builder](https://github.com/vmware-tanzu-labs/operator-builder)
  - [metacontroller](https://github.com/metacontroller/metacontroller)

- On doing some research we decided to focus on evaluating two of the most used frameworks i.e. Kubebuilder and operator-sdk. 

## Decision
- We built an operator POC using Kubebuilder as it is well maintained and it is also what most other frameworks(operator-sdk and operator-builder) rely on internally.
- We also found that moving from Kubebuilder to operator-sdk is an easier transition than vice-versa.
- We also didn't want to use operator-sdk as it's owned by a competitor.

## Future concerns
- Kubebuilder framework compatibility to list operator on operator hub.
- Validation of kube-builder based operator to run on Openshift and TMC
