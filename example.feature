@kubernetes
Feature: Matt
  Scenario: create one
    Given a resource called "namespace"
    """yaml
    apiVersion: v1
    kind: Namespace
    metadata:
      name: $NAMESPACE
    """
    And a resource called "pod"
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
    When I create namespace
    And I create pod
    Then in less than 3m pod's '{.status.conditions[?(@.type=="Ready")].status}' should equal "True"
