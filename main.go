package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"text/template"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	messages "github.com/cucumber/messages-go/v16"
	"github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"github.com/testernetes/bdk/format"
	"github.com/testernetes/bdk/kubernetes"
	"github.com/testernetes/gkube"
	"github.com/testernetes/trackedclient"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type VarsKey struct{}

// varsubRegex is the regular expression used to validate
// the var names before substitution
const varsubRegex = "^[_[:alpha:]][_[:alpha:][:digit:]]*$"

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "k8s",
}

func main() {
	godog.BindCommandLineFlags("", &opts)

	name := pflag.String("name", "bdk", "name")

	pflag.Parse()
	opts.Paths = pflag.Args()

	godog.Format("k8s", "Pretty Formatter for kubernetes", format.KubernetesFormatterFunc)

	testSuite := godog.TestSuite{
		Name:    *name,
		Options: &opts,
	}

	testSuite.ScenarioInitializer = func(sc *godog.ScenarioContext) {
		var cancel context.CancelFunc
		var sctx context.Context
		c := make(chan os.Signal, 1)

		sc.Before(func(ctx context.Context, s *godog.Scenario) (context.Context, error) {
			signal.Notify(c, os.Interrupt)
			sctx, cancel = context.WithCancel(ctx)

			go func() {
				select {
				case <-c:
					fmt.Printf("\nUser Interrupted, jumping to cleanup now. Press ^C again to skip cleanup.\n\n")
					cancel()
				case <-ctx.Done():
				}
			}()

			if isKubernetesScenario(s.Tags) {
				setupKubernetesScenarioRun(sc)
			}

			return sctx, nil
		})

		sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			signal.Stop(c)
			cancel()
			return ctx, nil
		})
		sc.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
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
				"getvar": getvar(ctx),
			}

			t := template.New("").Funcs(templateFuncs)

			// Process Text
			var err error
			st.Text, err = process(st.Text, t)
			if err != nil {
				return nil, err
			}

			// Process DocString & Table
			if st.Argument != nil {
				if st.Argument.DocString != nil {
					st.Argument.DocString.Content, err = process(st.Argument.DocString.Content, t)
					if err != nil {
						return nil, err
					}
				}
				// TODO Table
			}

			return ctx, nil
		})
		sc.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("User Interrupted")
			}
			return nil, nil
		})
	}

	testSuite.Options.ShowStepDefinitions = false

	os.Exit(testSuite.Run())
}

func process(s string, t *template.Template) (string, error) {
	t, err := t.Parse(s)
	if err != nil {
		return "", err
	}
	builder := &strings.Builder{}
	t.Execute(builder, nil)
	return builder.String(), nil
}

func getvar(ctx context.Context) func(key string) string {
	return func(key string) string {
		vars, ok := ctx.Value("variables").(map[string]string)
		if !ok {
			return ""
		}
		val, ok := vars[key]
		if !ok {
			return ""
		}
		return val
	}
}

func isKubernetesScenario(tags []*messages.PickleTag) bool {
	for t := range tags {
		if tags[t].Name == "@kubernetes" {
			return true
		}
	}
	return false
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
		gomega.Expect(objTrackingClient.DeleteAllTracked(context.Background())).Should(gomega.Succeed())
		return ctx, nil
	})
}
