package formatter

import (
	"strconv"
	"strings"
)

func parseFixtureActivity(line string) (fixtureDeclaration, bool) {
	parts := strings.Split(line, " -> ")
	if len(parts) != 2 {
		return fixtureDeclaration{}, false
	}
	left := strings.TrimPrefix(parts[0], "activity ")
	open := strings.Index(left, "(")
	close := strings.LastIndex(left, ")")
	if open <= 0 || close < open {
		return fixtureDeclaration{}, false
	}
	inputs := strings.TrimSpace(left[open+1 : close])
	var names []string
	if inputs != "" {
		for _, input := range strings.Split(inputs, ",") {
			names = append(names, strings.TrimSpace(input))
		}
	}
	return fixtureDeclaration{kind: ActivityDeclaration, name: left[:open], inputs: names, output: strings.TrimSpace(parts[1])}, true
}
func appendFixtureError(diagnostics Diagnostics, line int) Diagnostics {
	return append(diagnostics, Diagnostic{Severity: SeverityError, Code: CodeInvalidDocument, Message: "fixture parse error on line " + strconv.Itoa(line)})
}
