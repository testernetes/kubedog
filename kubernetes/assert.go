package kubernetes

import (
	"time"

	"github.com/cucumber/godog"
	. "github.com/matt-simons/gkube"
	"github.com/matt-simons/kubedog/assertion"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Could be generic assertion so we can get matcher related
type k8sAssertion struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Timeout  string `json:"timeout"`
	Interval string `json:"interval"`
}

func (a *k8sAssertion) GetUnstructured() *unstructured.Unstructured {
	Expect(a.TypeMeta.APIVersion).ShouldNot(BeEmpty(), "Provided test case resource has an empty API Version")
	Expect(a.TypeMeta.Kind).ShouldNot(BeEmpty(), "Provided test case resource has an empty Kind")
	Expect(a.ObjectMeta.Name).ShouldNot(BeEmpty(), "Provided test case resource has an empty Name")

	u := &unstructured.Unstructured{}
	u.SetAPIVersion(a.APIVersion)
	u.SetKind(a.Kind)
	u.SetName(a.Name)
	u.SetNamespace(a.Namespace)

	return u
}

func (a *k8sAssertion) GetTimeout() time.Duration {
	d, err := time.ParseDuration(a.Timeout)
	if err != nil {
		return 3 * time.Second
	}
	return d
}

func (a *k8sAssertion) GetInterval() time.Duration {
	d, err := time.ParseDuration(a.Interval)
	if err != nil {
		return 500 * time.Millisecond
	}
	return d
}

func (t *kubernetesScenario) AddAssertSteps(s *godog.ScenarioContext) {
	s.Step("^(eventually|consistently) `([^`]*)` should( not)? (.*)$", t.assert)
}

func (k *kubernetesScenario) assert(verb, jsonpath, not, matcherText string, manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	a := k.parseAssertion(manifest)

	u := a.GetUnstructured()
	m := WithJSONPath(jsonpath, assertion.GetMatcher(matcherText))

	timeout := a.GetTimeout()
	interval := a.GetInterval()

	switch {
	case verb == "eventually" && not == "":
		Eventually(k.Object(u)).WithTimeout(timeout).WithPolling(interval).Should(m)
	case verb == "eventually" && not == "not":
		Eventually(k.Object(u)).WithTimeout(timeout).WithPolling(interval).ShouldNot(m)
	case verb == "consistently" && not == "":
		Consistently(k.Object(u)).WithTimeout(timeout).WithPolling(interval).Should(m)
	case verb == "consistently" && not == "not":
		Consistently(k.Object(u)).WithTimeout(timeout).WithPolling(interval).ShouldNot(m)
	}

	return nil
}
