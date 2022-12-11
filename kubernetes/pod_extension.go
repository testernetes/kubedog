package kubernetes

import (
	"context"

	"github.com/cucumber/godog"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (k *kubernetesScenario) AddPodExtensionSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I execute "([^"]+)" in `+dns1123Name, k.iExecIn)
	sc.Step(`^I execute "([^"]+)" in `+dns1123Name+`/`+dns1123Name, k.iExecIn)
}

func (k *kubernetesScenario) iExecIn(ctx context.Context, cmd, ref string) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)
	Expect(u.GroupVersionKind().String()).Should(Equal("/v1, Kind=Pod"))

	pod := &corev1.Pod{}
	runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), pod)

	// Support writing to stdout/stderr similar to GinkgoWriter
	k.podSession, err = k.Exec(ctx, pod, "", []string{"/bin/sh", "-c", cmd}, nil, nil)
	Expect(err).ShouldNot(HaveOccurred())

	return nil
}
