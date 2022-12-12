package kubernetes

import (
	"io"
	"strings"

	"github.com/cucumber/godog"
	"github.com/testernetes/gkube"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/portforward"
)

type kubernetesScenario struct {
	gkube.KubernetesHelper
	objRegister       map[string]*unstructured.Unstructured
	podPortForwarders map[string]*portforward.PortForwarder
	podSessions       map[string]*gkube.PodSession

	out    io.Writer
	errOut io.Writer
}

func (k *kubernetesScenario) closePortForwarders() {
	for _, p := range k.podPortForwarders {
		p.Close()
	}
}

func NewKubernetesScenario(sc *godog.ScenarioContext, helper gkube.KubernetesHelper) kubernetesScenario {
	ks := kubernetesScenario{
		KubernetesHelper:  helper,
		objRegister:       make(map[string]*unstructured.Unstructured),
		podPortForwarders: make(map[string]*portforward.PortForwarder),
		podSessions:       make(map[string]*gkube.PodSession),
		out:               &strings.Builder{},
		errOut:            &strings.Builder{},
	}

	// Register Kubernetes Steps for the Scenario
	ks.AddCRUDSteps(sc)
	ks.AddPodExtensionSteps(sc)
	ks.AddAssertSteps(sc)
	ks.AddResourceSteps(sc)

	return ks
}
