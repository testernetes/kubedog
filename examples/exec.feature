@kubernetes
Feature: Matt
	Scenario: executing a script in a pod
		Given a resource called pod
		"""yaml
		apiVersion: v1
		kind: Pod
		metadata:
		  name: test
		  namespace: default
		spec:
		  containers:
		  - command:
		    - "sleep"
		    - "200"
		    image: busybox
		    name: bdk
		"""
		When I create pod
		Then within 1m pod's '{.status.conditions[?(@.type=="Ready")].status}' should equal "True"
		When I execute this script in pod:
		"""
		echo hello
		echo world
		"""
		Then pod should log "hello\nworld"
