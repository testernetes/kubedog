@kubernetes
Feature: Matt
  Scenario: create one
    Given I create a resource:
    """yaml
    apiVersion: v1
    kind: Namespace
    metadata:
      name: $NAMESPACE
    """
    And I create a resource:
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
    Then eventually `{.status.conditions[?(@.type=="Ready")].status}` should equal "True"
    """yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: test
      namespace: $NAMESPACE
    timeout: 2m
    """
