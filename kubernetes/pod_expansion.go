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
	sc.Step(`^I evict "([^"]+)"`, k.iEvict)
	sc.Step(`^I execute "([^"]+)" in `+dns1123Name, k.iExecInDefaultContainer)
	sc.Step(`^I execute this script in `+dns1123Name, k.iExecScriptInDefaultContainer)
	sc.Step(`^I execute this script in `+dns1123Name, k.iExecScriptInContainer)
	sc.Step(`^I execute "([^"]+)" in `+dns1123Name+`/`+dns1123Name, k.iExecInContainer)
	sc.Step(`^I port forward `+dns1123Name+`on ports ([0-9: ]+)`, k.iPortForwardPod)
}

func (k *kubernetesScenario) iEvict(ctx context.Context, ref string) (err error) {
	defer failHandler(&err)
	pod := k.getPodFromRegister(ref)

	Eventually(k.Evict).WithContext(ctx).WithArguments(pod).Should(Succeed())

	return nil
}

func (k *kubernetesScenario) iExecScriptInDefaultContainer(ctx context.Context, ref string, script *godog.DocString) (err error) {
	return k.iExecScriptInContainer(ctx, ref, "", script)
}

func (k *kubernetesScenario) iExecScriptInContainer(ctx context.Context, ref, container string, script *godog.DocString) (err error) {
	defer failHandler(&err)

	pod := k.getPodFromRegister(ref)

	cmd := script.Content
	shell := script.MediaType

	if shell == "" {
		shell = "/bin/sh"
	}

	session, err := k.Exec(ctx, pod, container, []string{shell, "-c", cmd}, k.out, k.errOut)
	Expect(err).ShouldNot(HaveOccurred())
	k.podSessions[ref] = session

	return nil
}

func (k *kubernetesScenario) iExecInDefaultContainer(ctx context.Context, cmd, ref string) (err error) {
	return k.iExecInContainer(ctx, cmd, ref, "")
}

func (k *kubernetesScenario) iExecInContainer(ctx context.Context, cmd, ref, container string) (err error) {
	defer failHandler(&err)

	pod := k.getPodFromRegister(ref)

	session, err := k.Exec(ctx, pod, container, []string{"/bin/sh", "-c", cmd}, k.out, k.errOut)
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
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), pod)

	Expect(err).ShouldNot(HaveOccurred(), "reference was not a pod")
	return pod
}
