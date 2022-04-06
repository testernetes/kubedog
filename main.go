package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	messages "github.com/cucumber/messages-go/v16"
	"github.com/drone/envsubst"
	"github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"github.com/testernetes/gkube"
	"github.com/testernetes/kubedog/format"
	"github.com/testernetes/kubedog/kubernetes"
	"github.com/testernetes/trackedclient"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// varsubRegex is the regular expression used to validate
// the var names before substitution
const varsubRegex = "^[_[:alpha:]][_[:alpha:][:digit:]]*$"

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "k8s",
}

func main() {
	godog.BindCommandLineFlags("", &opts)

	name := pflag.String("name", "kubedog", "name")

	pflag.Parse()
	opts.Paths = pflag.Args()

	godog.Format("k8s", "Pretty Formatter for kubernetes", format.KubernetesFormatterFunc)

	testSuite := godog.TestSuite{
		Name:    *name,
		Options: &opts,
	}

	testSuite.ScenarioInitializer = func(sc *godog.ScenarioContext) {
		sc.BeforeScenario(func(s *godog.Scenario) {
			if isKubernetesScenario(s.Tags) {
				setupKubernetesScenarioRun(sc)
			}
		})
		sc.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
			vars, ok := ctx.Value(&kubernetes.VarsKey{}).(map[string]string)
			if !ok || len(vars) == 0 {
				return ctx, nil
			}
			err := stepVariableSubstitution(st, vars)
			return nil, err
		})
	}

	testSuite.Options.ShowStepDefinitions = false

	os.Exit(testSuite.Run())
}

func isKubernetesScenario(tags []*messages.PickleTag) bool {
	for t := range tags {
		if tags[t].Name == "@kubernetes" {
			return true
		}
	}
	return false
}

func stepVariableSubstitution(step *godog.Step, vars map[string]string) error {
	text, err := sub(step.Text, vars)
	if err != nil {
		return err
	}
	step.Text = text

	if step.Argument != nil {
		if step.Argument.DocString != nil {
			content, err := sub(step.Argument.DocString.Content, vars)
			if err != nil {
				return err
			}
			step.Argument.DocString.Content = content
		}
	}
	return nil
}

func sub(in string, vars map[string]string) (string, error) {
	// run bash variable substitutions
	out, err := envsubst.Eval(in, func(s string) string {
		return vars[s]
	})
	if err != nil {
		return out, fmt.Errorf("variable substitution failed: %w", err)
	}
	return out, nil
}

func setupKubernetesScenarioRun(sc *godog.ScenarioContext) {
	gomega.RegisterFailHandler(func(message string, _ ...int) {
		panic(message)
	})

	// Setup client
	cfg := config.GetConfigOrDie()
	opts := client.Options{
		Scheme: clientgoscheme.Scheme,
	}

	objTrackingClient, err := trackedclient.New(cfg, opts)
	if err != nil {
		panic(err)
	}

	kubernetes.NewKubernetesScenario(sc, gkube.NewKubernetesHelper(gkube.WithClient(objTrackingClient)))

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		gomega.Expect(objTrackingClient.DeleteAllTracked(ctx)).Should(gomega.Succeed())
		return ctx, nil
	})
}
