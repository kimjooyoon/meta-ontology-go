package bodycodegen

import (
	"strings"
	"testing"
)

func TestGenerateLowersLetAssignmentAndConditionalReturn(t *testing.T) {
	source := []byte(`package bodycodegen
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity ClampBelowZero(Integer) -> Integer computes "let value = input\nvalue = input\nlet accepted = value >= 0\nif accepted { return value } else { return 0 }"
`)
	result, err := Generate("main.gooo", source, "ClampBelowZero")
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.Decision != "PASS" || !result.Report.TypecheckPassed || !result.Report.DeterministicReplay {
		t.Fatalf("report does not prove accepted lowering: %#v", result.Report)
	}
	if result.Report.CompletenessPercent != 100 || result.Report.SourceConstructs != result.Report.LoweredConstructs {
		t.Fatalf("accepted constructs are not fully accounted for: %#v", result.Report)
	}
	for _, want := range []string{
		`//gooo:generated:start id="bodycodegen://activity/clamp-below-zero" kind="activity"`,
		"func ClampBelowZero(input int64) int64",
		"var value = input",
		"value = input",
		"if accepted {",
	} {
		if !strings.Contains(result.Source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, result.Source)
		}
	}
	if result.Report.GeneratedDigest != result.Report.ReplayDigest || result.Report.RepositoryWrites != 0 {
		t.Fatalf("output replay or write boundary is incorrect: %#v", result.Report)
	}
}

func TestGenerateFailsClosedForUnsupportedOrIncompletePrograms(t *testing.T) {
	tests := []struct {
		name, program string
	}{
		{name: "call", program: `return helper(input)`},
		{name: "missing else", program: `if input > 0 { return input } return 0`},
		{name: "missing return", program: `let value = input`},
		{name: "type mismatch", program: `return input >= 0`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := []byte("package bodycodegen\nnamespace bodycodegen\nentity Integer id \"bodycodegen://entity/integer\"\nactivity Clamp(Integer) -> Integer computes \"" + test.program + "\"\n")
			if result, err := Generate("main.gooo", source, "Clamp"); err == nil || result.Source != "" {
				t.Fatalf("unsupported program produced output: result=%#v err=%v", result, err)
			}
		})
	}
}
