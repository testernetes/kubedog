package kubernetes

import (
	"context"
	"strings"

	"github.com/cucumber/godog"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (k *kubernetesScenario) AddPodExtensionSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I execute "([^"]+)" in `+dns1123Name, k.iExecIn)
	sc.Step(`^I execute "([^"]+)" in `+dns1123Name+`/`+dns1123Name, k.iExecIn)
	sc.Step(`^I port forward `+dns1123Name+`on ports ([0-9: ]+)`, k.iPortForwardPod)
}

func (k *kubernetesScenario) iExecIn(ctx context.Context, cmd, ref string) (err error) {
	defer failHandler(&err)

	pod := k.getPodFromRegister(ref)

	session, err := k.Exec(ctx, pod, "", []string{"/bin/sh", "-c", cmd}, k.out, k.errOut)
	Expect(err).ShouldNot(HaveOccurred())
	k.podSessions[ref] = session

	return nil
}

func (k *kubernetesScenario) iPortForwardPod(ctx context.Context, ref, ports string) (err error) {
	defer failHandler(&err)

	pod := k.getPodFromRegister(ref)

	forwardedPorts := strings.Split(ports, " ")

	pf, err := k.PortForward(ctx, pod, forwardedPorts, k.out, k.errOut)
	Expect(err).ShouldNot(HaveOccurred())
	k.podPortForwarders[ref] = pf

	return nil
}

func (k *kubernetesScenario) getPodFromRegister(ref string) *corev1.Pod {
	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)
	Expect(u.GroupVersionKind().String()).Should(Equal("/v1, Kind=Pod"))

	pod := &corev1.Pod{}
	runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), pod)
	return pod
}
