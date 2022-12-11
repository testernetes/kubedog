package kubernetes

import (
	"context"
	"fmt"
	"os"

	"github.com/cucumber/godog"
	. "github.com/onsi/gomega"
)

func (k *kubernetesScenario) AddResourceSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the following variables:$`, k.variables)
	sc.Step(`^a resource called `+dns1123Name, k.AResource)
	sc.Step(`^a resource called `+dns1123Name+` from file (.*)`, k.AResourceFromFile)
	sc.Step(`^the following resources$`, k.resources)
}

func (k *kubernetesScenario) variables(ctx context.Context, table *godog.Table) (context.Context, error) {

	vars, ok := ctx.Value("variables").(map[string]string)
	if !ok {
		vars = make(map[string]string)
	}

	// assert header
	// | key | value

	for _, row := range table.Rows {
		if len(row.Cells) > 2 {
			return ctx, fmt.Errorf("invalid table")
		}
		k, v := row.Cells[0].Value, row.Cells[1].Value
		vars[k] = v
	}

	ctx = context.WithValue(ctx, "variables", vars)

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
