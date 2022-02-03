package kubernetes

import (
	"fmt"

	"github.com/cucumber/godog"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func (k *kubernetesScenario) parseResource(manifest *godog.DocString) *unstructured.Unstructured {
	Expect(manifest.MediaType).Should(BeElementOf("json", "yaml"), "Unrecognised content-type %s. Supported types are json, yaml.", manifest.MediaType)

	u := &unstructured.Unstructured{}
	Expect(yaml.Unmarshal([]byte(manifest.Content), u)).Should(Succeed())

	ns := u.GetNamespace()
	u.SetNamespace(ns)
	return u
}

func (t *kubernetesScenario) parseAssertion(manifest *godog.DocString) *k8sAssertion {
	Expect(manifest.MediaType).Should(BeElementOf("json", "yaml"), "Unrecognised content-type %s. Supported types are json, yaml.", manifest.MediaType)

	a := &k8sAssertion{}
	Expect(yaml.Unmarshal([]byte(manifest.Content), a)).Should(Succeed())
	return a
}

func failHandler(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("%s", r)
	}
}
