package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

const ansiEscape = "\x1b"

// a color code type
type color int

// some ansi colors
const (
	black color = iota + 30
	red
	green
	yellow
	blue
	magenta
	cyan
	white
)

func (f *kubernetesFmt) printStepResult(c colors.ColorFunc, s string, a ...interface{}) {
	s = strings.Repeat(" ", 4) + s
	s = c(s)
	fmt.Fprintf(f.out, s, a...)
}

func (f *kubernetesFmt) print(c color, i int, s string, a ...interface{}) {
	s = strings.Repeat(" ", i) + s
	//s = strings.ReplaceAll(s, "\n", "\n"+strings.Repeat(" ", i))
	s = fmt.Sprintf("%s[%dm%v%s[0m", ansiEscape, c, s, ansiEscape)
	fmt.Fprintf(f.out, s, a...)
}

func KubernetesFormatterFunc(suite string, out io.Writer) godog.Formatter {
	return newKubernetesFmt(suite, out)
}

func newKubernetesFmt(suite string, out io.Writer) *kubernetesFmt {
	return &kubernetesFmt{
		PrettyFmt: godog.NewPrettyFmt(suite, out),
		out:       out,
	}
}

type kubernetesFmt struct {
	*godog.PrettyFmt

	out io.Writer
}

func (f *kubernetesFmt) Snippets(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.out.Write([]byte("nothing"))
}

func (f *kubernetesFmt) Summary() {
	undefinedStepResults := f.Storage.MustGetPickleStepResultsByStatus(godog.StepUndefined)
	if len(undefinedStepResults) > 0 {
		f.print(yellow, 0, "\n--- Invalid steps:\n")

		for _, undef := range undefinedStepResults {
			step := f.Storage.MustGetPickleStep(undef.PickleStepID)
			f.print(yellow, 0, "Invalid Step: %s", step.Text)
		}
		fmt.Fprintln(f.out)
	}

	failedStepResults := f.Storage.MustGetPickleStepResultsByStatus(godog.StepFailed)
	if len(failedStepResults) > 0 {
		f.print(red, 0, "\n--- Failed steps:\n")

		for _, fail := range failedStepResults {
			scenario := f.Storage.MustGetPickle(fail.PickleID)
			step := f.Storage.MustGetPickleStep(fail.PickleStepID)
			feature := f.Storage.MustGetFeature(scenario.Uri)

			astScenario := feature.FindScenario(scenario.AstNodeIds[0])
			astStep := feature.FindStep(step.AstNodeIds[0])

			// indent multiline errors
			failureError := strings.ReplaceAll(fmt.Sprintf("%s", fail.Err), "\n", "\n      ")

			f.print(red, 0, "%s %s\n", feature.Feature.Keyword, feature.Feature.Name)
			f.print(red, 2, "%s: %s\n", astScenario.Keyword, scenario.Name)
			f.print(red, 4, "%s%s\n", astStep.Keyword, step.Text)
			f.print(red, 6, "%s\n", failureError)
		}
		fmt.Fprintln(f.out)
	}
}

func (f *kubernetesFmt) Passed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.printStep(scenario, step)
}

func (f *kubernetesFmt) Skipped(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.printStep(scenario, step)
}

func (f *kubernetesFmt) Undefined(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.printStep(scenario, step)
}

func (f *kubernetesFmt) Failed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition, err error) {
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.printStep(scenario, step)
}

//func (f *kubernetesFmt) Pending(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
//	f.PrettyFmt.Base.Pending(scenario, step, match)
//}

func (f *kubernetesFmt) printStep(scenario *godog.Scenario, step *godog.Step) {
	feature := f.Storage.MustGetFeature(scenario.Uri)
	astBackground := feature.FindBackground(scenario.AstNodeIds[0])
	astScenario := feature.FindScenario(scenario.AstNodeIds[0])
	astStep := feature.FindStep(step.AstNodeIds[0])

	var astBackgroundStep bool
	var firstExecutedBackgroundStep bool
	//var backgroundSteps int
	if astBackground != nil {
		//backgroundSteps = len(astBackground.Steps)

		for idx, backgroundStep := range astBackground.Steps {
			if step.Id == backgroundStep.Id {
				astBackgroundStep = true
				firstExecutedBackgroundStep = idx == 0
				break
			}
		}
	}

	firstPickle := feature.Pickles[0].Id == scenario.Id

	if astBackgroundStep && !firstPickle {
		return
	}

	if astBackgroundStep && firstExecutedBackgroundStep {
		fmt.Fprintln(f.out)
		f.print(white, 2, "%s: %s\n", astBackground.Keyword, astBackground.Name)
	}

	//if !astBackgroundStep && len(astScenario.Examples) > 0 {
	//	f.printOutlineExample(scenario, backgroundSteps)
	//	return
	//}

	//scenarioHeaderLength, maxLength := f.scenarioLengths(scenario)
	//stepLength := f.lengthPickleStep(astStep.Keyword, step.Text)

	firstExecutedScenarioStep := astScenario.Steps[0].Id == step.AstNodeIds[0]
	if !astBackgroundStep && firstExecutedScenarioStep {
		fmt.Fprintln(f.out)
		f.print(white, 2, "%s: %s\n", astScenario.Keyword, astScenario.Name)
	}

	stepResult := f.Storage.MustGetPickleStepResult(step.Id)
	stepColor := stepResult.Status.Color()
	f.printStepResult(stepColor, "%s%s\n", astStep.Keyword, astStep.Text)

	//if step.Argument != nil {
	//	if table := step.Argument.DataTable; table != nil {
	//		f.printTable(table, cyan)
	//	}

	//	if docString := astStep.DocString; docString != nil {
	//		f.printDocString(docString)
	//	}
	//}

	if stepResult.Err != nil {
		stepErr := strings.ReplaceAll(stepResult.Err.Error(), "\n", "\n      ")
		f.printStepResult(stepColor, "%+v\n", stepErr)
	}
}
