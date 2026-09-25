package languageutility

import (
	"strings"
	"testing"
)

func TestGeneratedMetaProgramContainsEveryUtilityCell(t *testing.T) {
	program, err := GenerateProgram(fixtureContract())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(program, "activity Observe") != 42 {
		t.Fatalf("observed cell activities = %d", strings.Count(program, "activity Observe"))
	}
	if !strings.Contains(program, "activity MeasureLanguageUtility") ||
		!strings.Contains(program, "gooo.utility.debugging.resource-observed.regression:v1") {
		t.Fatalf("generated program is not utility-bound:\n%s", program)
	}
}
