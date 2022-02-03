package assertion

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"sigs.k8s.io/yaml"
)

var (
	matcherEqual     = regexp.MustCompile(`^equal (.*)$`)
	matcherElementOf = regexp.MustCompile(`^be an element of (.*)$`)
	matcherConsistOf = regexp.MustCompile(`^consist of (.*)$`)
	matcherNumeric   = regexp.MustCompile(`^(?:be )?([~=<>]{1,2}) (\d+)$`)
	matcherBool      = regexp.MustCompile(`^be (true|false)$`)
	matcherContain   = regexp.MustCompile(`^contain (.*)$`)
	matcherPrefix    = regexp.MustCompile(`^have prefix (.*)$`)
	matcherSuffix    = regexp.MustCompile(`^have suffix (.*)$`)
	matcherRegex     = regexp.MustCompile(`^match regex (.*)$`)
	matcherLen       = regexp.MustCompile(`^have length (\d)$`)
)

func getWords(in string) (out []interface{}) {
	scanner := bufio.NewScanner(strings.NewReader(in))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		var word interface{}
		Expect(yaml.Unmarshal(scanner.Bytes(), &word)).Should(Succeed())
		out = append(out, word)
	}
	return
}

func GetMatcher(text string) types.GomegaMatcher {
	switch {
	case matcherEqual.MatchString(text):
		expectedBytes := []byte(matcherEqual.FindStringSubmatch(text)[1])
		var expected interface{}
		Expect(yaml.Unmarshal(expectedBytes, &expected)).Should(Succeed())
		return BeEquivalentTo(expected)

	case matcherRegex.MatchString(text):
		expected := matcherRegex.FindStringSubmatch(text)[1]
		return MatchRegexp(string(expected))

	case matcherLen.MatchString(text):
		fields := matcherLen.FindStringSubmatch(text)
		expected, err := strconv.Atoi(string(fields[1]))
		Expect(err).ShouldNot(HaveOccurred(), "Length must be a positive integer")
		return HaveLen(expected)

	case matcherContain.MatchString(text):
		expected := matcherContain.FindStringSubmatch(text)[1]
		return ContainSubstring(string(expected))

	case matcherPrefix.MatchString(text):
		expected := matcherContain.FindStringSubmatch(text)[1]
		return HavePrefix(string(expected))

	case matcherSuffix.MatchString(text):
		expected := matcherContain.FindStringSubmatch(text)[1]
		return HaveSuffix(string(expected))

	case matcherNumeric.MatchString(text):
		fields := matcherNumeric.FindSubmatch([]byte(text))
		expectedBytes := fields[2]
		var expected interface{}
		Expect(yaml.Unmarshal(expectedBytes, &expected)).Should(Succeed())
		comparator := string(fields[1])
		return BeNumerically(comparator, expected)

	case matcherBool.MatchString(text):
		fields := matcherBool.FindStringSubmatch(string(text))
		if fields[1] == "true" {
			return BeTrue()
		}
		return BeFalse()

	case matcherElementOf.MatchString(text):
		fields := matcherElementOf.FindStringSubmatch(string(text))
		return BeElementOf(getWords(fields[1])...)

	case matcherConsistOf.MatchString(text):
		fields := matcherConsistOf.FindStringSubmatch(string(text))
		return ConsistOf(getWords(fields[1])...)

	default:
		panic("unrecognised assertion")
	}
}
