package kubernetes

import (
	"context"

	"github.com/cucumber/godog"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func (k *kubernetesScenario) AddCRUDSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I create `+dns1123Name, k.iCreateAResource)
	sc.Step(`^I patch `+dns1123Name, k.iPatchAResource)
	sc.Step(`^I delete `+dns1123Name, k.iDeleteAResource)
}

func (k *kubernetesScenario) iCreateAResource(ctx context.Context, ref string) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)

	Eventually(k.Create).WithContext(ctx).WithArguments(u).Should(Succeed())

	return nil
}

func (k *kubernetesScenario) iPatchAResource(ctx context.Context, ref string, manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)

	patch, err := yaml.YAMLToJSON([]byte(manifest.Content))
	Expect(err).ShouldNot(HaveOccurred(), patchContentErrMsg, err)
	Eventually(k.Patch).WithContext(ctx).WithArguments(u, client.RawPatch(types.StrategicMergePatchType, patch)).Should(Succeed())

	return nil
}

func (k *kubernetesScenario) iDeleteAResource(ctx context.Context, ref string, manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)
	Eventually(k.Delete).WithContext(ctx).WithArguments(u).Should(Succeed())

	return nil
}
