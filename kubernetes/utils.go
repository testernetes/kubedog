package kubernetes

import (
	"bytes"
	"fmt"
	"runtime/debug"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func (k *kubernetesScenario) parseResource(r []byte) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	r = bytes.ReplaceAll(r, []byte("\t"), []byte("  "))
	Expect(yaml.Unmarshal(r, u)).Should(Succeed())

	Expect(u.GetAPIVersion()).ShouldNot(BeEmpty(), noAPIVersionErrMsg)
	Expect(u.GetKind()).ShouldNot(BeEmpty(), noKindErrMsg)
	Expect(u.GetName()).ShouldNot(BeEmpty(), noNameErrMsg)

	return u
}

func failHandler(err *error) {
	// set a debugging flag to enable stack traces or not
	if r := recover(); r != nil {
		*err = fmt.Errorf("%s\n\n%s", r, string(debug.Stack()))
		//*err = fmt.Errorf("%s", r)
	}
}
