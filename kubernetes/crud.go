package kubernetes

import (
	"github.com/cucumber/godog"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const dns1123Name = "([a-z0-9]([-a-z0-9]*[a-z0-9])?)"

const (
	noResourceErrMsg   string = "No resource called %s was registered in a previous step"
	patchContentErrMsg string = "Error while reading patch contents: %s"
	noAPIVersionErrMsg string = "Provided test case resource has an empty API Version"
	noKindErrMsg       string = "Provided test case resource has an empty Kind"
	noNameErrMsg       string = "Provided test case resource has an empty Name"
)

func (k *kubernetesScenario) AddCRUDSteps(sc *godog.ScenarioContext) {
	sc.Step(`^a resource called `+dns1123Name, k.AResource)
	sc.Step(`^I patch `+dns1123Name, k.iPatchAResource)
	sc.Step(`^I create `+dns1123Name, k.iCreateAResource)
	sc.Step(`^I delete `+dns1123Name, k.iDeleteAResource)
}

func (k *kubernetesScenario) AResource(reference string, manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	u := k.parseResource(manifest)
	Expect(u).ShouldNot(BeNil())

	Expect(u.GetAPIVersion()).ShouldNot(BeEmpty(), noAPIVersionErrMsg)
	Expect(u.GetKind()).ShouldNot(BeEmpty(), noKindErrMsg)
	Expect(u.GetName()).ShouldNot(BeEmpty(), noNameErrMsg)

	k.objRegister[reference] = u

	return nil
}

func (k *kubernetesScenario) iCreateAResource(ref string) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)
	Eventually(k.Create(u)).Should(Succeed())

	return nil
}

func (k *kubernetesScenario) iPatchAResource(ref string, manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)

	patch, err := yaml.YAMLToJSON([]byte(manifest.Content))
	Expect(err).ShouldNot(HaveOccurred(), patchContentErrMsg, err)
	Eventually(k.Patch(u, client.RawPatch(types.StrategicMergePatchType, patch))).Should(Succeed())

	return nil
}

func (k *kubernetesScenario) iDeleteAResource(ref string, manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)
	Eventually(k.Delete(u)).Should(Succeed())

	return nil
}
