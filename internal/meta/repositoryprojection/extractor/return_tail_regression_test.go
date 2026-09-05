package extractor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func TestReturnTailNamedRegressionSet(t *testing.T) {
	cases := []struct {
		name string
		kind string
		src  string
	}{
		{name: "R1 rendered-header-overflow-admission", kind: "pass", src: returnTailHeaderOverflowFixture()},
		{name: "R2 oversized-whole-tail-rejection-termination", kind: "refuted", src: returnTailWholeTailFixture()},
		{name: "R3 typed-nil-conversion", kind: "pass", src: returnTailTypedNilFixture()},
		{name: "R4 unknown-callee-UNKNOWN", kind: "unknown", src: returnTailUnknownCalleeFixture()},
		{name: "R5 local-callee-pointer-write-not-pure", kind: "unknown", src: returnTailPointerCalleeFixture()},
		{name: "R6 parenthesized-exposed-value-mutation", kind: "refuted", src: returnTailExposedMutationFixture()},
		{name: "R7 incomplete-callee-typed-evidence-UNKNOWN", kind: "unknown", src: returnTailIncompleteCalleeFixture()},
		{name: "R8 IR-dependency-tampering", kind: "contract", src: ""},
		{name: "R9 generated-receipt-unit-binding", kind: "binding", src: returnTailTypedNilFixture()},
	}
	if len(cases) != 9 {
		t.Fatalf("named regression denominator=%d, want 9", len(cases))
	}
	for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
			if tc.kind == "contract" {
				values := returnTailContractBaseline()
				values[1].InputEntity = "FunctionInput"
				if _, ok := returnTailContractObligations(values); ok {
					t.Fatal("tampered obligation dependency was admitted")
				}
				return
			}
			result, err := extractReturnTailRegressionFixture(t, tc.src)
			if tc.kind == "pass" {
				if err != nil || len(result.Evidence) == 0 {
					t.Fatalf("regression pass failed: result=%+v err=%v", result, err)
				}
				if tc.name == "R3 typed-nil-conversion" && !generatedContains(result.Generated, "(*typedError)(nil)") {
					t.Fatal("typed nil expression was not preserved")
				}
				return
			}
			if tc.kind == "binding" {
				if err != nil || len(result.Evidence) != 1 || result.Evidence[0].Operation != "extract-function" || result.Evidence[0].ContractActivity != "ExtractFunction" || result.Evidence[0].ContractOutputEntity != "OperationResult" || len(result.Evidence[0].ContractObligations) != 6 || len(result.Generated) < 2 {
					t.Fatalf("generated unit binding evidence=%+v generated=%d err=%v", result.Evidence, len(result.Generated), err)
				}
				return
			}
			var failure Failure
			if !errors.As(err, &failure) {
				t.Fatalf("regression failure=%v", err)
			}
			if tc.kind == "refuted" && failure.UnknownClass != "KNOWN_CONTRADICTION" {
				t.Fatalf("failure=%+v, want confirmed contradiction", failure)
			}
			if tc.kind == "unknown" && failure.UnknownClass != "DIRECT_MISSING" {
				t.Fatalf("failure=%+v, want unproven evidence", failure)
			}
		})
	}
}

func returnTailContractBaseline() []generation.OperationInputContractObligationEvidence {
	values := make([]generation.OperationInputContractObligationEvidence, 0, len(returnTailObligations))
	previous := "FunctionInput"
	for _, name := range returnTailObligations {
		activity, output := returnTailContractStage(name)
		values = append(values, generation.OperationInputContractObligationEvidence{Name: name, Activity: activity, InputEntity: previous, OutputEntity: output, UsedInputFact: true, GeneratedOutputFact: true})
		previous = output
	}
	return values
}

func extractReturnTailRegressionFixture(t *testing.T, source string) (Result, error) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return ExtractWithResult(root, "x.go")
}

func generatedContains(generated map[string][]byte, wanted string) bool {
	for _, data := range generated {
		if strings.Contains(string(data), wanted) {
			return true
		}
	}
	return false
}

func returnTailHeaderOverflowFixture() string {
	return "package p\n\nfunc F(values map[string]struct{}) error {\n" + strings.Repeat("\t_ = 1\n", 69) + "\tif len(values) != 0 {\n\t\treturn nil\n\t}\n\treturn nil\n}\n\nfunc errorSentinel() error { return &sentinelError{} }\n\ntype sentinelError struct{ error }\n"
}

func returnTailWholeTailFixture() string {
	return "package p\n\nfunc F(values map[string]struct{}) error {\n" + strings.Repeat("\t_ = 1\n", 4) + "\tif len(values) != 0 {\n" + strings.Repeat("\t\t_ = 1\n", 70) + "\t\treturn nil\n\t}\n\treturn nil\n}\n"
}

func returnTailTypedNilFixture() string {
	return "package p\n\ntype typedError struct{ error }\n\nfunc F() error {\n" + strings.Repeat("\t_ = 1\n", 72) + "\treturn (*typedError)(nil)\n}\n"
}

func returnTailUnknownCalleeFixture() string {
	return "package p\n\nimport \"runtime\"\n\nfunc F() error {\n" + strings.Repeat("\t_ = 1\n", 72) + "\t_, _, _, _ = runtime.Caller(0)\n\treturn nil\n}\n"
}

func returnTailPointerCalleeFixture() string {
	return "package p\n\ntype Box struct{ Field error }\n\nfunc mutate(box *Box) { box.Field = nil }\n\nfunc F() error {\n\tbox := &Box{}\n" + strings.Repeat("\t_ = 1\n", 72) + "\tmutate(box)\n\treturn nil\n}\n\nfunc errorSentinel() error { return &sentinelError{} }\n\ntype sentinelError struct{ error }\n"
}

func returnTailExposedMutationFixture() string {
	return "package p\n\nvar exposed *Box\n\ntype Box struct{ Field error }\n\nfunc F() error {\n\tx := Box{}\n\texposed = &(x)\n" + strings.Repeat("\t_ = 1\n", 72) + "\t(x).Field = errorSentinel()\n\treturn nil\n}\n\nfunc errorSentinel() error { return &sentinelError{} }\n\ntype sentinelError struct{ error }\n"
}

func returnTailIncompleteCalleeFixture() string {
	return "package p\n\nfunc F() error {\n" + strings.Repeat("\t_ = 1\n", 72) + "\treturn missingCallee()\n}\n"
}
