package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/assertion"
	. "github.com/testernetes/gkube"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (k *kubernetesScenario) AddAssertSteps(sc *godog.ScenarioContext) {
	eventuallyPhrases := []string{
		"in less than ",
		"in under ",
		"in no more than ",
		"at least ",
		"within ",
		"",
	}
	for _, phrase := range eventuallyPhrases {
		sc.Step(fmt.Sprintf(`^%s(\d+\w{1,2})*[,]?\s?([a-z0-9][-a-z0-9]*[a-z0-9])'s '([^']*)' should (.*)$`, phrase), k.eventuallyObjectWithTimeout)
		sc.Step(fmt.Sprintf(`^%s(\d+\w{1,2})*[,]?\s?([a-z0-9][-a-z0-9]*[a-z0-9])'s '([^']*)' should not (.*)$`, phrase), k.eventuallyNotObjectWithTimeout)
		sc.Step(fmt.Sprintf(`^%s(\d+\w{1,2})*[,]?\s?([a-z0-9][-a-z0-9]*[a-z0-9])'s exit code should be (\d+)$`, phrase), k.exitCodeShouldBe)
		sc.Step(fmt.Sprintf(`^%s(\d+\w{1,2})*[,]?\s?([a-z0-9][-a-z0-9]*[a-z0-9]) should log "([^"]*)"$`, phrase), k.shouldSay)
	}
	consistentlyPhrases := []string{
		"for at least",
		"for no less than",
	}
	for _, phrase := range consistentlyPhrases {
		sc.Step(fmt.Sprintf(`^%s (\w*)[,]? ([a-z0-9][-a-z0-9]*[a-z0-9])'s '([^']*)' should (.*)$`, phrase), k.consistentlyObjectWithTimeout)
		sc.Step(fmt.Sprintf(`^%s (\w*)[,]? ([a-z0-9][-a-z0-9]*[a-z0-9])'s '([^']*)' should not (.*)$`, phrase), k.consistentlyNotObjectWithTimeout)
	}
}

func (k *kubernetesScenario) eventuallyObjectWithTimeout(ctx context.Context, timeout, ref, jsonpath, matcherText string) (err error) {
	defer failHandler(&err)
	o, matcher, d := k.parseAssertion(ref, jsonpath, matcherText, timeout)
	ctx, _ = context.WithTimeout(ctx, d)
	Eventually(k.Object).WithContext(ctx).WithArguments(o).WithTimeout(d).Should(matcher)
	return nil
}

func (k *kubernetesScenario) eventuallyNotObjectWithTimeout(ctx context.Context, timeout, ref, jsonpath, matcherText string) (err error) {
	defer failHandler(&err)
	o, matcher, d := k.parseAssertion(ref, jsonpath, matcherText, timeout)
	ctx, _ = context.WithTimeout(ctx, d)
	Eventually(k.Object).WithContext(ctx).WithArguments(o).WithTimeout(d).ShouldNot(matcher)
	return nil
}

func (k *kubernetesScenario) consistentlyObjectWithTimeout(ctx context.Context, timeout, ref, jsonpath, matcherText string) (err error) {
	defer failHandler(&err)
	o, matcher, d := k.parseAssertion(ref, jsonpath, matcherText, timeout)
	ctx, _ = context.WithTimeout(ctx, d)
	Consistently(k.Object).WithContext(ctx).WithArguments(o).WithTimeout(d).Should(matcher)
	return nil
}

func (k *kubernetesScenario) consistentlyNotObjectWithTimeout(ctx context.Context, timeout, ref, jsonpath, matcherText string) (err error) {
	defer failHandler(&err)
	o, matcher, d := k.parseAssertion(ref, jsonpath, matcherText, timeout)
	ctx, _ = context.WithTimeout(ctx, d)
	Consistently(k.Object).WithContext(ctx).WithArguments(o).WithTimeout(d).ShouldNot(matcher)
	return nil
}

func (k *kubernetesScenario) exitCodeShouldBe(ctx context.Context, timeout, ref string, code int) (err error) {
	defer failHandler(&err)

	d, err := time.ParseDuration(timeout)
	Expect(err).ShouldNot(HaveOccurred())

	s, ok := k.podSessions[ref]
	Expect(ok).Should(BeTrue())
	Eventually(s).WithTimeout(d).Should(Exit(code))

	return nil
}

func (k *kubernetesScenario) shouldSay(ctx context.Context, timeout, ref, message string) (err error) {
	defer failHandler(&err)

	d, err := time.ParseDuration(timeout)
	Expect(err).ShouldNot(HaveOccurred())

	s, ok := k.podSessions[ref]
	Expect(ok).Should(BeTrue())
	Eventually(s).WithTimeout(d).Should(Say(message))

	return nil
}

func (k *kubernetesScenario) parseAssertion(ref, jsonpath, matcherText, timeout string) (*unstructured.Unstructured, types.GomegaMatcher, time.Duration) {
	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)

	if timeout == "" {
		timeout = "1s"
	}
	d, err := time.ParseDuration(timeout)
	Expect(err).ShouldNot(HaveOccurred())

	return u, HaveJSONPath(jsonpath, assertion.GetMatcher(matcherText)), d
}
