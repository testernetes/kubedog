package kubernetes

import (
	"fmt"
	"time"

	"github.com/cucumber/godog"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	. "github.com/tkube/gkube"
	"github.com/tkube/kubedog/assertion"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Could be generic assertion so we can get matcher related
type k8sAssertion struct {
	Timeout  string `json:"timeout"`
	Interval string `json:"interval"`
}

func (k *kubernetesScenario) AddAssertSteps(s *godog.ScenarioContext) {
	eventuallyPhrases := []string{
		"in less than",
		"in under",
		"in no more than",
	}
	for _, phrase := range eventuallyPhrases {
		s.Step(fmt.Sprintf(`^%s (\w+) ([a-z0-9][-a-z0-9]*[a-z0-9])'s '([^']*)' should (.*)$`, phrase), k.eventuallyObjectWithTimeout)
		s.Step(fmt.Sprintf(`^%s (\w+) ([a-z0-9][-a-z0-9]*[a-z0-9])'s '([^']*)' should not (.*)$`, phrase), k.eventuallyNotObjectWithTimeout)
	}
	consistentlyPhrases := []string{
		"for at least",
		"for no less than",
	}
	for _, phrase := range consistentlyPhrases {
		s.Step(fmt.Sprintf(`^%s (\w+) ([a-z0-9][-a-z0-9]*[a-z0-9])'s '([^']*)' should (.*)$`, phrase), k.consistentlyObjectWithTimeout)
		s.Step(fmt.Sprintf(`^%s (\w+) ([a-z0-9][-a-z0-9]*[a-z0-9])'s '([^']*)' should not (.*)$`, phrase), k.consistentlyNotObjectWithTimeout)
	}
}

func (k *kubernetesScenario) eventuallyObjectWithTimeout(timeout, ref, jsonpath, matcherText string) (err error) {
	o, matcher, d := k.parseAssertion(ref, jsonpath, matcherText, timeout)
	Eventually(k.Object(o)).WithTimeout(d).Should(matcher)
	return nil
}

func (k *kubernetesScenario) eventuallyNotObjectWithTimeout(timeout, ref, jsonpath, matcherText string) (err error) {
	o, matcher, d := k.parseAssertion(ref, jsonpath, matcherText, timeout)
	Eventually(k.Object(o)).WithTimeout(d).ShouldNot(matcher)
	return nil
}

func (k *kubernetesScenario) consistentlyObjectWithTimeout(timeout, ref, jsonpath, matcherText string) (err error) {
	o, matcher, d := k.parseAssertion(ref, jsonpath, matcherText, timeout)
	Consistently(k.Object(o)).WithTimeout(d).Should(matcher)
	return nil
}

func (k *kubernetesScenario) consistentlyNotObjectWithTimeout(timeout, ref, jsonpath, matcherText string) (err error) {
	o, matcher, d := k.parseAssertion(ref, jsonpath, matcherText, timeout)
	Consistently(k.Object(o)).WithTimeout(d).ShouldNot(matcher)
	return nil
}

func (k *kubernetesScenario) parseAssertion(ref, jsonpath, matcherText, timeout string) (*unstructured.Unstructured, types.GomegaMatcher, time.Duration) {
	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceError, ref)

	d, err := time.ParseDuration(timeout)
	Expect(err).ShouldNot(HaveOccurred())

	return u, WithJSONPath(jsonpath, assertion.GetMatcher(matcherText)), d
}
