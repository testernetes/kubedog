# Behaviour Driven Kubernetes

A BDD testing framework for Kubernetes using Syntax.

Enables assertion testing of Kubernetes workloads without needing progamming knowledge.

## Scenarios and Steps

Scenarios are defined by Given, When, and Then steps and must be in defined in that order and never repeating:
* Given steps should describe the existing environment setup
* When steps should execute an action.
* Then steps should assert that the expected state is met.
* And and But steps can optionally be used where multiple Givens, Whens, or Thens are needed

## Assertions

An assertion can be made against either a jsonpath, container log, or port.

[Step] [optional time constraint] [resource reference] [jsonpath/log/port] should/should not [assertion] [expected]

## Examples

In the following example a pod resource is defined in the `Given` step.
The resource is given a name called 'mypod' which can be referenced in future steps in the same Scenario.

```feature
@kubernetes
Feature: Creating resources
  Scenario: create a busybox pod that sleeps
    Given a resource called mypod
    """yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: test
      namespace: $NAMESPACE
    spec:
      containers:
      - command:
        - sleep
        - "200"
        image: busybox:latest
        name: busybox
    """
    When I create mypod
    Then in less than 3m pod's '{.status.conditions[?(@.type=="Ready")].status}' should equal "True"
```
