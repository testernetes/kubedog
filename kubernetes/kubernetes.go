package kubernetes

import (
	"github.com/cucumber/godog"
	"github.com/matt-simons/gkube"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type kubernetesScenario struct {
	gkube.KubernetesHelper
	objRegister map[string]*unstructured.Unstructured
}

func NewKubernetesScenario(sc *godog.ScenarioContext, helper gkube.KubernetesHelper) kubernetesScenario {
	ks := kubernetesScenario{
		KubernetesHelper: helper,
		objRegister:      make(map[string]*unstructured.Unstructured),
	}

	// Register Kubernetes Steps for the Scenario
	ks.AddCRUDSteps(sc)
	ks.AddAssertSteps(sc)

	return ks
}
