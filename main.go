package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	messages "github.com/cucumber/messages-go/v16"
	"github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"github.com/tkube/gkube"
	"github.com/tkube/kubedog/kubernetes"
	"github.com/tkube/trackedclient"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
}

func main() {
	godog.BindCommandLineFlags("", &opts)

	name := pflag.String("name", "kubedog", "name")

	pflag.Parse()
	opts.Paths = pflag.Args()

	testSuite := godog.TestSuite{
		Name:    *name,
		Options: &opts,
	}

	testSuite.ScenarioInitializer = func(sc *godog.ScenarioContext) {
		sc.BeforeScenario(func(s *godog.Scenario) {
			if isKubernetesScenario(s.Tags) {
				replaceVariables(s)
				setupKubernetesScenarioRun(sc)
			}
		})
	}

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

func replaceVariables(s *godog.Scenario) {
	uniqueNamespace := "test-" + uuidOrDie()
	for i := range s.Steps {
		step := s.Steps[i]
		step.Text = strings.ReplaceAll(step.Text, "$NAMESPACE", uniqueNamespace)
		if step.Argument != nil {
			step.Argument.DocString.Content = strings.ReplaceAll(step.Argument.DocString.Content, "$NAMESPACE", uniqueNamespace)
		}
	}
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
		objTrackingClient.DeleteAllTracked(ctx)
		return ctx, err
	})
}

func uuidOrDie() string {
	u := make([]byte, 16)
	_, err := rand.Read(u)
	if err != nil {
		panic(err)
	}

	u[8] = (u[8] | 0x80) & 0xBF
	u[6] = (u[6] | 0x40) & 0x4F

	return hex.EncodeToString(u)
}
