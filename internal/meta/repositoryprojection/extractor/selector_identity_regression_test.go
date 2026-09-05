package extractor

import (
	"errors"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestFirstOversizedFunctionUsesActualDeclarationIdentity(t *testing.T) {
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		t.Fatal(err)
	}
	policy := sourcepolicy.Default()
	if contract.Operation != sourcepolicy.OperationExtractFunction ||
		contract.Activity != "ExtractFunction" ||
		contract.InputEntity != "FunctionInput" ||
		contract.OutputEntity != "OperationResult" ||
		contract.InputSubjectKind != sourcepolicy.SubjectKindFunction ||
		!contract.UsedInputFact || !contract.GeneratedOutputFact ||
		contract.SourceDigest == "" || contract.SemanticDigest == "" ||
		!sourcepolicy.IsLineCapMetric(sourcepolicy.DimensionFunctionLines) ||
		policy.MaxFunctionLines != functionLineLimit {
		t.Fatalf("ExtractFunction/source-policy selector contract is incomplete: contract=%+v policy=%+v", contract, policy)
	}
	t.Logf("extract-function-contract operation=%s activity=%s input=%s output=%s subject_kind=%s metric=%s function_line_limit=%d source_digest=%s semantic_digest=%s",
		contract.Operation, contract.Activity, contract.InputEntity, contract.OutputEntity,
		contract.InputSubjectKind, sourcepolicy.DimensionFunctionLines, policy.MaxFunctionLines,
		contract.SourceDigest, contract.SemanticDigest)

	cases := []struct {
		name       string
		source     string
		wantMethod bool
		wantNil    bool
	}{
		{
			name: "short-method-long-free-target",
			source: "package p\n\ntype T struct{}\n\nfunc (T) F() {}\n\nfunc F() {\n" +
				strings.Repeat("\t_ = 1\n", functionLineLimit+1) + "}\n",
		},
		{
			name: "oversized-method-unsafe-boundary",
			source: "package p\n\ntype T struct{}\n\nfunc (T) F() {\n" +
				strings.Repeat("\t_ = 1\n", functionLineLimit+1) + "}\n\nfunc F() {}\n",
			wantMethod: true,
		},
		{
			name: "short-method-long-doc-helper",
			source: "package p\n\ntype T struct{}\n\n" +
				strings.Repeat("// method documentation\n", functionLineLimit+1) + "func (T) F() {}\n\nfunc F() {}\n",
			wantMethod: true,
		},
		{
			name:    "small-method-only",
			source:  "package p\n\ntype T struct{}\n\nfunc (T) F() {}\n",
			wantNil: true,
		},
	}
	if len(cases) != 4 {
		t.Fatalf("selector identity regression cohort denominator=%d, want 4", len(cases))
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.source) > 4096 {
				t.Fatalf("fixture exceeds bounded regression input: %d bytes", len(tc.source))
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "fixture.go", []byte(tc.source), parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			function := firstOversizedFunction(fset, file, []byte(tc.source))
			if tc.wantNil {
				if function != nil {
					t.Fatalf("small method-only fixture selected %s", functionIdentity(fset, function))
				}
				t.Logf("extract-function metric=%s selected=false method_only=true contract_source_digest=%s contract_semantic_digest=%s",
					sourcepolicy.DimensionFunctionLines, contract.SourceDigest, contract.SemanticDigest)
				return
			}
			if function == nil {
				t.Fatal("oversized fixture selected no declaration")
			}
			if (function.Recv != nil) != tc.wantMethod {
				t.Fatalf("selected declaration=%s method=%t, want method=%t", functionIdentity(fset, function), function.Recv != nil, tc.wantMethod)
			}
			helper, helperErr := renderedDeclarationHelper(fset, file, []byte(tc.source), function)
			if helperErr != nil {
				t.Fatalf("selected declaration helper render failed: %v", helperErr)
			}
			t.Logf("extract-function metric=%s selected_declaration=%s selected_function_span_lines=%d selected_helper_physical_lines=%d method=%t contract_source_digest=%s contract_semantic_digest=%s",
				sourcepolicy.DimensionFunctionLines, functionIdentity(fset, function), declarationLines(fset, function), physicalLines(helper), function.Recv != nil,
				contract.SourceDigest, contract.SemanticDigest)
			if !tc.wantMethod {
				return
			}
			_, _, err = decomposeFunction(t.TempDir(), "fixture.go", []byte(tc.source), fset, file, function)
			var failure Failure
			if !errors.As(err, &failure) || failure.Reason != "METHOD_SUFFIX_DECOMPOSITION_UNSAFE" {
				t.Fatalf("oversized method decomposition=%v, want METHOD_SUFFIX_DECOMPOSITION_UNSAFE", err)
			}
			t.Logf("method-boundary=selected actual method remains unsafe reason=%s", failure.Reason)
		})
	}
}
