@kubernetes
Feature: Matt
	Scenario: create one
		Given the following variables:
			| namespace | matt |
		Given a resource called namespace
		"""yaml
		apiVersion: v1
		kind: Namespace
		metadata:
		  name: {{ getvar "namespace" }}
		"""
		And a resource called pod
		"""yaml
		apiVersion: v1
		kind: Pod
		metadata:
		  name: test
		  namespace: {{ getvar "namespace" }}
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
		Then within 1m pod's '{.status.conditions[?(@.type=="Ready")].status}' should equal "True"
		When I execute "echo hello" in pod
		Then within 20s pod should log "hello"
