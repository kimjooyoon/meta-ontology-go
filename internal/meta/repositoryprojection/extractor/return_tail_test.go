package extractor

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestReturnTailSafetyMatrix(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		positive bool
	}{
		{name: "positive terminal error tail", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tif len(values) != 0 {\n\t\treturn nil\n\t}\n\treturn nil\n"), positive: true},
		{name: "positive early return is preserved", source: returnTailPrefixBindingFixture("func F(values map[string]struct{}) error {\n", "\tif len(values) == 0 {\n\t\treturn errorSentinel()\n\t}\n", "\tif len(values) != 0 {\n\t\treturn nil\n\t}\n\treturn nil\n"), positive: true},
		{name: "named result", source: returnTailFixture("func F(values map[string]struct{}) (err error) {\n", "\tif len(values) != 0 {\n\t\treturn err\n\t}\n\treturn err\n"), positive: false},
		{name: "method", source: returnTailFixture("func (T) F(values map[string]struct{}) error {\n", "\tif len(values) != 0 {\n\t\treturn nil\n\t}\n\treturn nil\n"), positive: false},
		{name: "go statement", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tgo func() {}()\n\treturn nil\n"), positive: false},
		{name: "defer statement", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tdefer func() {}()\n\treturn nil\n"), positive: false},
		{name: "escaping branch", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tgoto done\n\tdone:\n\treturn nil\n"), positive: false},
		{name: "address escape stale pointer", source: returnTailPrefixBindingFixture("func F(values map[string]struct{}) error {\n", "\terr := error(nil)\n\tp := &err\n\t_ = p\n", "\t*p = errorSentinel()\n\treturn err\n"), positive: false},
		{name: "closure capture stale copy", source: returnTailPrefixBindingFixture("func F(values map[string]struct{}) error {\n", "\terr := error(nil)\n\tset := func() { err = errorSentinel() }\n\t_ = set\n", "\tset()\n\treturn err\n"), positive: false},
		{name: "false helper capacity proof", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tif len(values) != 0 {\n"+strings.Repeat("\t\t_ = 1\n", 70)+"\t\treturn nil\n\t}\n\treturn nil\n"), positive: false},
	}
	if len(cases) != 10 {
		t.Fatalf("safety matrix denominator=%d, want 10", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(tc.source), 0o644); err != nil {
				t.Fatal(err)
			}
			result, err := ExtractWithResult(root, "x.go")
			if tc.positive {
				if err != nil {
					t.Fatalf("positive case failed: %v", err)
				}
				if len(result.Evidence) != 1 || result.Evidence[0].Strategy != returnTailStrategy {
					t.Fatalf("strategy evidence=%+v", result.Evidence)
				}
				if len(result.Evidence[0].Obligations) != len(returnTailObligations) {
					t.Fatalf("obligations=%+v", result.Evidence[0].Obligations)
				}
				var selectedPreflight *PreflightObservationEvidence
				for index := range result.Evidence[0].PreflightObservations {
					observation := &result.Evidence[0].PreflightObservations[index]
					if observation.Subject == "func:F" {
						selectedPreflight = observation
						break
					}
				}
				if selectedPreflight == nil || selectedPreflight.Activity != "ExtractFunction" || selectedPreflight.Metric != sourcepolicy.DimensionFunctionLines ||
					selectedPreflight.HelperMeasurementScope != renderedCapacityHelperMeasurementScope ||
					selectedPreflight.FunctionStatus != string(renderedCapacityOverCap) || selectedPreflight.SourceDigest == "" ||
					selectedPreflight.ContractSourceDigest == "" || selectedPreflight.ContractSemanticDigest == "" {
					t.Fatalf("preflight evidence=%+v, want selected function observation with bound digests", result.Evidence[0].PreflightObservations)
				}
				finalCapacity := result.Evidence[0].FinalRenderedCapacity
				if result.Evidence[0].BeforeFunctionLines <= functionLineLimit || result.Evidence[0].AfterFunctionLines > functionLineLimit ||
					result.Evidence[0].RenderedHelperLines > functionLineLimit ||
					result.Evidence[0].BeforeRenderedCapacityOverage <= result.Evidence[0].AfterRenderedCapacityOverage || result.Evidence[0].AfterRenderedCapacityOverage < 0 ||
					finalCapacity == nil || finalCapacity.Scope != "final-generated-functions" || finalCapacity.Lines <= 0 || finalCapacity.Overage != 0 {
					t.Fatalf("capacity evidence=%+v", result.Evidence[0])
				}
				for path, data := range result.Generated {
					if physicalLines(data) > functionLineLimit {
						t.Fatalf("generated unit %s exceeds capacity: %d lines", path, physicalLines(data))
					}
				}
				if !generatedFunctionContains(result.Generated, "F", "return FExtractedReturnTail") {
					t.Fatal("outer function did not use a return-valued helper")
				}
				if tc.name == "positive early return is preserved" && !generatedFunctionContains(result.Generated, "F", "return errorSentinel()") {
					t.Fatal("outer early return was not preserved")
				}
				return
			}
			var failure Failure
			if !errors.As(err, &failure) {
				t.Fatalf("negative case error=%v", err)
			}
			if tc.name == "closure capture stale copy" {
				// A refuted candidate does not refute another candidate with unknown effects.
				if failure.Reason != "CALLEE_EFFECTS_UNPROVEN" || failure.UnknownClass != "DIRECT_MISSING" {
					t.Fatalf("whole-search closure evidence=%+v error=%v", failure, err)
				}
				assertReturnTailClosureCaptureRejected(t, tc.source)
				return
			}
			if failure.Reason != "NO_SAFE_DECLARATION_CAPACITY" && failure.Reason != "METHOD_SUFFIX_DECOMPOSITION_UNSAFE" {
				t.Fatalf("negative case reason=%s error=%v", failure.Reason, err)
			}
		})
	}
}

func TestReturnTailPreparationProgressDefersFinalCapacityProof(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := returnTailFixture("func F(values map[string]struct{}) error {\n", "\tif len(values) != 0 {\n\t\treturn nil\n\t}\n\treturn nil\n")
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	_, evidence, err := prepareOversizedFunctions(root, "x.go", []byte(source), fset, file)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range evidence {
		if item.Strategy != returnTailStrategy || item.AfterRenderedCapacityOverage <= 0 {
			continue
		}
		if item.PreparationProgress == nil || len(item.ProofStages) != len(returnTailObligations)-2 || len(item.Obligations) != len(returnTailObligations)-2 ||
			item.FinalRenderedCapacity != nil {
			t.Fatalf("return-tail intermediate evidence=%+v, want progress without final capacity proof", item)
		}
		return
	}
	t.Fatalf("preparation evidence=%+v, want a return-tail intermediate progress record", evidence)
}

func assertReturnTailClosureCaptureRejected(t *testing.T, source string) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	function, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok || function.Name.Name != "F" || function.Body == nil || len(function.Body.List) == 0 {
		t.Fatal("closure fixture lacks the target function")
	}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}, Types: map[ast.Expr]types.TypeAndValue{}}
	configuration := types.Config{}
	if _, err := configuration.Check("fixture", fset, []*ast.File{file}, info); err != nil {
		t.Fatal(err)
	}
	statements := function.Body.List[len(function.Body.List)-1:]
	bindings, err := suffixBindings(statements, function, fset, typeEvidence{info: info})
	if err != nil || len(bindings) != 1 || bindings[0].name != "err" {
		t.Fatalf("terminal candidate free bindings=%+v error=%v", bindings, err)
	}
	err = hasReturnTailBindingHazard(function.Body, statements, bindings, info)
	if !isKnownSuffixContradiction(err) {
		t.Fatalf("captured free binding was not refuted: error=%v", err)
	}
}

func generatedFunctionContains(generated map[string][]byte, name, wanted string) bool {
	needle := "func " + name + "("
	for _, data := range generated {
		text := string(data)
		if strings.Contains(text, needle) && strings.Contains(text, wanted) {
			return true
		}
	}
	return false
}

func returnTailFixture(header, tail string) string {
	return "package p\n\n" + header + strings.Repeat("\t_ = 1\n", 72) + tail + "}\n\nfunc errorSentinel() error { return &sentinelError{} }\n\ntype sentinelError struct{}\n\nfunc (*sentinelError) Error() string { return \"sentinel\" }\n\ntype T struct{}\n"
}

func returnTailPrefixBindingFixture(header, prefix, tail string) string {
	return "package p\n\n" + header + prefix + strings.Repeat("\t_ = 1\n", 72) + tail + "}\n\nfunc errorSentinel() error { return &sentinelError{} }\n\ntype sentinelError struct{}\n\nfunc (*sentinelError) Error() string { return \"sentinel\" }\n\ntype T struct{}\n"
}
