package kubernetes

import (
	"github.com/cucumber/godog"
	. "github.com/onsi/gomega"
)

func (k *kubernetesScenario) AddCreateSteps(s *godog.ScenarioContext) {
	s.Step(`^I create a resource:$`, k.iCreateAResource)
	s.Step(`^I delete a resource:$`, k.iDeleteAResource)
}

func (k *kubernetesScenario) iCreateAResource(manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	u := k.parseResource(manifest)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(u).ShouldNot(BeNil())

	Eventually(k.Create(u)).Should(Succeed())

	return nil
}

func (k *kubernetesScenario) iDeleteAResource(manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	u := k.parseResource(manifest)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(u).ShouldNot(BeNil())

	Eventually(k.Delete(u)).Should(Succeed())

	return nil
}
