package kubernetes

import (
	"github.com/cucumber/godog"
	"github.com/matt-simons/gkube"
)

type kubernetesScenario struct {
	gkube.KubernetesHelper
}

func NewKubernetesScenario(sc *godog.ScenarioContext, helper gkube.KubernetesHelper) kubernetesScenario {
	ks := kubernetesScenario{
		helper,
	}

	// Register Kubernetes Steps for the Scenario
	ks.AddCreateSteps(sc)
	ks.AddAssertSteps(sc)

	return ks
}
