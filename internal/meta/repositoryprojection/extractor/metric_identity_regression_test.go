package extractor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestNamedFunctionLinesFreeFunctionIdentityRegressionCohort(t *testing.T) {
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
		t.Fatalf("ExtractFunction/source-policy function-span contract is incomplete: contract=%+v policy=%+v", contract, policy)
	}
	t.Logf("extract-function-contract operation=%s activity=%s input=%s output=%s subject_kind=%s metric=%s function_line_limit=%d source_digest=%s semantic_digest=%s",
		contract.Operation, contract.Activity, contract.InputEntity, contract.OutputEntity,
		contract.InputSubjectKind, sourcepolicy.DimensionFunctionLines, policy.MaxFunctionLines,
		contract.SourceDigest, contract.SemanticDigest)

	cases := []struct {
		name   string
		source string
		found  bool
	}{
		{
			name: "method-short/free-long",
			source: "package p\n\ntype T struct{}\n\nfunc (T) F() {}\n\nfunc F() {\n" +
				strings.Repeat("\t_ = 1\n", functionLineLimit+1) + "}\n",
			found: true,
		},
		{
			name: "method-long/free-short",
			source: "package p\n\ntype T struct{}\n\nfunc (T) F() {\n" +
				strings.Repeat("\t_ = 1\n", functionLineLimit+1) + "}\n\nfunc F() {}\n",
			found: true,
		},
		{
			name: "method-only",
			source: "package p\n\ntype T struct{}\n\nfunc (T) F() {}\n",
			found: false,
		},
	}
	if len(cases) != 3 {
		t.Fatalf("function identity regression cohort denominator=%d, want 3", len(cases))
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.source) > 4096 {
				t.Fatalf("fixture exceeds bounded regression input: %d bytes", len(tc.source))
			}
			wantLines, wantFound := freeFunctionSpan(t, tc.source, "F")
			if wantFound != tc.found {
				t.Fatalf("fixture free-function shape=%t, want %t", wantFound, tc.found)
			}
			gotLines, gotFound := namedFunctionLines([]byte(tc.source), "F")
			if gotFound != wantFound || gotLines != wantLines {
				t.Fatalf("named function span=%d found=%t, want free-function span=%d found=%t", gotLines, gotFound, wantLines, wantFound)
			}
			t.Logf("extract-function metric=%s free_function_span_lines=%d found=%t contract_source_digest=%s contract_semantic_digest=%s",
				sourcepolicy.DimensionFunctionLines, gotLines, gotFound, contract.SourceDigest, contract.SemanticDigest)
		})
	}
}

func freeFunctionSpan(t *testing.T, source, name string) (int, bool) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Name == nil || function.Name.Name != name {
			continue
		}
		return declarationLines(fset, function), true
	}
	return 0, false
}
