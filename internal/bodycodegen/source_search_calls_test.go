package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func searchCallsSource(helper, body, grammar, cases string) []byte {
	return []byte(fmt.Sprintf(`package searchcalls
namespace searchcalls
entity Integer id "searchcalls://integer"
entity Box id "searchcalls://box" fields { field value id "searchcalls://value" type integer required one }
%s
activity Build(Integer) -> Integer computes %q assembling {
 search hole "value" grammar %q intent "Compose a source-declared helper with a searched expression." max_candidates "16"
 %s
 attempts "16"
}
`, helper, body, grammar, cases))
}

func TestSourceIRSearchPureCalls(t *testing.T) {
	for _, tc := range []struct {
		name, helper, body, grammar, cases string
		probes                             int
	}{
		{"scalar", `activity Step(Integer) -> Integer computes "return input + 1"`,
			"return Step(__GOOO_BODY_HOLE_value__)", "integer-offset-constant/v1",
			`case "2" -> "3" case "10" -> "11" holdout_case "100" -> "101"`, 0},
		{"record residual", `activity Step(Integer) -> Box computes "return Box{value: input * 2}"`,
			"let result = Step(input + __GOOO_BODY_HOLE_value__); return result.value", "integer-hole-residual/v1",
			`case "2" -> "6" case "10" -> "22" holdout_case "100" -> "202"`, 4},
		{"record argument", `activity Step(Integer, Box) -> Integer computes "return input0 + input1.value"`,
			"return Step(input, Box{value: __GOOO_BODY_HOLE_value__})", "integer-offset-constant/v1",
			`case "2" -> "4" case "10" -> "20" holdout_case "100" -> "200"`, 0},
		{"quadratic", `activity Step(Integer) -> Integer computes "return input * input"`,
			"return input + Step(__GOOO_BODY_HOLE_value__)", "integer-hole-quadratic/v1",
			`case "2" -> "11" case "10" -> "19" holdout_case "100" -> "109"`, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			source := searchCallsSource(tc.helper, tc.body, tc.grammar, tc.cases)
			spec, err := SourceAssembly(ctx, "calls.gooo", source, "Build")
			if err != nil {
				t.Fatal(err)
			}
			if err := ValidateSourceIRSearch(ctx, "calls.gooo", source, "Build", spec); err != nil {
				t.Fatal("preflight lost the source helper", err)
			}
			result, err := GenerateWithSourceIRSearch(ctx, "calls.gooo", source, "Build", spec, "", "")
			if err != nil {
				t.Fatal(err)
			}
			r := result.Report.BodySearch
			if r.TrainingPassed != 2 || r.HoldoutPassed != 1 || r.ProviderOperations != 0 ||
				result.Report.CallClosure == nil || len(result.Report.CallClosure.Activities) != 1 ||
				result.Report.CallClosure.Activities[0].Name != "Step" {
				t.Fatal("search or its emitted helper differs", r)
			}
			if tc.probes > 0 && (r.CandidateGeneration.HoleContext == nil || r.CandidateGeneration.HoleContext.EvaluationCalls != tc.probes) {
				t.Fatal("source helper was absent from residual probes", r.CandidateGeneration)
			}
			if err := VerifyIRBodySearchProjection(ctx, "calls.gooo", source, result); err != nil {
				t.Fatal("source helper was absent from replay", err)
			}
			changed := []byte(strings.Replace(string(source), tc.helper, strings.Replace(tc.helper, "input", "(input + 1)", 1), 1))
			if err := VerifyIRBodySearchProjection(ctx, "calls.gooo", changed, result); err == nil {
				t.Fatal("changed helper source accepted with old evidence")
			}
		})
	}
}

func TestSourceIRSearchPureCallRejections(t *testing.T) {
	for _, helper := range []string{"", `activity Step(Integer) -> Integer computes "return Step(input)"`} {
		source := searchCallsSource(helper, "return Step(__GOOO_BODY_HOLE_value__)",
			"integer-offset-constant/v1", `case "2" -> "2"`)
		spec, err := SourceAssembly(context.Background(), "invalid.gooo", source, "Build")
		if err == nil {
			err = ValidateSourceIRSearch(context.Background(), "invalid.gooo", source, "Build", spec)
		}
		if err == nil {
			t.Fatal("unknown or recursive helper accepted")
		}
	}
}
