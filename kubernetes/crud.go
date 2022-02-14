package kubernetes

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/cucumber/godog"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type podSessionKey struct{}
type VarsKey struct{}

//const dns1123Name = "([a-z0-9]+[-a-z0-9]*[a-z0-9])"
const dns1123Name = `(\w+)`

// varsubRegex is the regular expression used to validate
// the var names before substitution
const varsubRegex = "^[_[:alpha:]][_[:alpha:][:digit:]]*$"

const (
	noResourceErrMsg   string = "No resource called %s was registered in a previous step"
	patchContentErrMsg string = "Error while reading patch contents: %s"
	noAPIVersionErrMsg string = "Provided test case resource has an empty API Version"
	noKindErrMsg       string = "Provided test case resource has an empty Kind"
	noNameErrMsg       string = "Provided test case resource has an empty Name"
	notPodErrMsg       string = "Provided resource is not a Pod"
)

func (k *kubernetesScenario) AddCRUDSteps(sc *godog.ScenarioContext) {
	sc.Step(`^(\w+)=(.*)$`, k.setVariable)
	sc.Step(`^a resource called `+dns1123Name, k.AResource)
	sc.Step(`^a resource called `+dns1123Name+` from file (.*)`, k.AResourceFromFile)
	sc.Step(`^the following resources$`, k.resources)
	sc.Step(`^I patch `+dns1123Name, k.iPatchAResource)
	sc.Step(`^I create `+dns1123Name, k.iCreateAResource)
	sc.Step(`^I delete `+dns1123Name, k.iDeleteAResource)
	sc.Step(`^I execute "([^"]+)" in `+dns1123Name, k.iExecIn)
	sc.Step(`^I execute "([^"]+)" in `+dns1123Name+`/`+dns1123Name, k.iExecIn)
}

var templateFuncs = template.FuncMap{
	"rand": func() string {
		u := make([]byte, 16)
		_, err := rand.Read(u)
		if err != nil {
			panic(err)
		}

		u[8] = (u[8] | 0x80) & 0xBF
		u[6] = (u[6] | 0x40) & 0x4F

		return hex.EncodeToString(u)
	},
	"getenv": os.Getenv,
}

func (k *kubernetesScenario) setVariable(ctx context.Context, key, val string) (context.Context, error) {
	vars, ok := ctx.Value(&VarsKey{}).(map[string]string)
	if !ok {
		vars = make(map[string]string)
	}

	r, _ := regexp.Compile(varsubRegex)
	if !r.MatchString(key) {
		return ctx, fmt.Errorf("'%s' var name is invalid, must match '%s'", key, varsubRegex)
	}

	// val is a golang template
	t, err := template.New(key).Funcs(templateFuncs).Parse(val)
	if err != nil {
		return ctx, err
	}
	newVal := &strings.Builder{}
	t.Execute(newVal, nil)
	vars[key] = newVal.String()

	ctx = context.WithValue(ctx, &VarsKey{}, vars)

	return ctx, nil
}

func (k *kubernetesScenario) AResourceFromFile(ref, file string) (err error) {
	defer failHandler(&err)

	content, err := os.ReadFile(file)
	k.objRegister[ref] = k.parseResource(content)

	return nil
}

func (k *kubernetesScenario) AResource(ref string, manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	Expect(manifest.MediaType).Should(BeElementOf("json", "yaml"), "Unrecognised content-type %s. Supported types are json, yaml.", manifest.MediaType)
	k.objRegister[ref] = k.parseResource([]byte(manifest.Content))

	return nil
}

func (k *kubernetesScenario) resources(resources *godog.Table) (err error) {
	defer failHandler(&err)

	Expect(len(resources.Rows)).Should(BeNumerically(">", 1))

	for i := 1; i < len(resources.Rows); i++ {
		row := resources.Rows[i]
		Expect(row.Cells).Should(HaveLen(2))

		ref := row.Cells[0].Value
		file := row.Cells[1].Value

		data, err := os.ReadFile(file)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(data).ShouldNot(BeEmpty())
		u := k.parseResource(data)

		k.objRegister[ref] = u
	}
	return nil
}

func (k *kubernetesScenario) iCreateAResource(ref string) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)
	Eventually(k.Create(u)).Should(Succeed())

	return nil
}

func (k *kubernetesScenario) iExecIn(cmd, ref string) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)
	Expect(u.GroupVersionKind().String()).Should(Equal("/v1, Kind=Pod"))

	pod := &corev1.Pod{}
	runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), pod)

	k.podSession, err = k.Exec(pod, &corev1.PodExecOptions{
		Command: []string{"/bin/sh", "-c", cmd},
		Stdout:  true,
		Stderr:  true,
	}, time.Minute, nil, nil)
	Expect(err).ShouldNot(HaveOccurred())

	return nil
}

func (k *kubernetesScenario) iPatchAResource(ref string, manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)

	patch, err := yaml.YAMLToJSON([]byte(manifest.Content))
	Expect(err).ShouldNot(HaveOccurred(), patchContentErrMsg, err)
	Eventually(k.Patch(u, client.RawPatch(types.StrategicMergePatchType, patch))).Should(Succeed())

	return nil
}

func (k *kubernetesScenario) iDeleteAResource(ref string, manifest *godog.DocString) (err error) {
	defer failHandler(&err)

	u, ok := k.objRegister[ref]
	Expect(ok).Should(BeTrue(), noResourceErrMsg, ref)
	Eventually(k.Delete(u)).Should(Succeed())

	return nil
}
