package extractor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestExtractFunctionHelperProjectionRegressionCohort(t *testing.T) {
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		t.Fatal(err)
	}
	if contract.Operation != sourcepolicy.OperationExtractFunction ||
		contract.Activity != "ExtractFunction" ||
		contract.InputEntity != "FunctionInput" ||
		contract.OutputEntity != "OperationResult" ||
		contract.InputSubjectKind != sourcepolicy.SubjectKindFunction ||
		!contract.UsedInputFact || !contract.GeneratedOutputFact ||
		contract.SourceDigest == "" || contract.SemanticDigest == "" {
		t.Fatalf("ExtractFunction input identity is incomplete: %+v", contract)
	}

	cases := []struct {
		name   string
		source string
		fn     string
	}{
		{
			name:   "plain-free-function",
			source: "package fixture\n\nfunc F() {\n\tvalue := 1\n\t_ = value\n}\n",
			fn:     "F",
		},
		{
			name:   "selected-imports-and-comments",
			source: "package fixture\n\nimport (\n\t\"encoding/json\"\n\t\"strconv\"\n)\n\n// F retains the selected declaration documentation.\nfunc F() {\n\tvar _ json.RawMessage\n\t_ = strconv.IntSize\n}\n",
			fn:     "F",
		},
		{
			name:   "free-function-over-same-named-method",
			source: "package fixture\n\ntype T struct{}\n\nfunc (T) F() error { return nil }\n\nfunc F() error { return nil }\n",
			fn:     "F",
		},
	}
	if len(cases) != 3 {
		t.Fatalf("helper projection cohort denominator=%d, want 3", len(cases))
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.source) > 4096 {
				t.Fatalf("fixture exceeds bounded regression input: %d bytes", len(tc.source))
			}
			helperOnly, fullRenderer, err := renderHelperFixture(tc.source, tc.fn)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(helperOnly, fullRenderer) {
				t.Fatalf("helper-only output differs from full renderer:\nhelper-only=%q\nfull=%q", helperOnly, fullRenderer)
			}
			t.Logf("helper-byte-equivalence=helper-only bytes equal full-renderer bytes=%d", len(helperOnly))
			t.Logf("extract-function-contract operation=%s activity=%s input=%s output=%s subject_kind=%s source_digest=%s semantic_digest=%s",
				contract.Operation, contract.Activity, contract.InputEntity, contract.OutputEntity,
				contract.InputSubjectKind, contract.SourceDigest, contract.SemanticDigest)

			fset := token.NewFileSet()
			if _, err := parser.ParseFile(fset, "helper.go", helperOnly, parser.ParseComments); err != nil {
				t.Fatalf("helper-only syntax assertion failed: %v", err)
			}
			t.Log("syntax-assertion=helper-only output parses")
		})
	}
}

func renderHelperFixture(source, name string) ([]byte, []byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", []byte(source), parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	list, err := imports(file)
	if err != nil {
		return nil, nil, err
	}
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Name == nil || function.Name.Name != name {
			continue
		}
		start := function.Pos()
		if function.Doc != nil {
			start = function.Doc.Pos()
		}
		selected := []declaration{{node: function, start: fset.Position(start).Offset, end: fset.Position(function.End()).Offset, identity: "func:" + name}}
		full, err := render(fset, file, []byte(source), selected, list)
		if err != nil {
			return nil, nil, err
		}
		helperOnly, err := renderedFunctionHelper([]byte(source), name)
		if err != nil {
			return nil, nil, err
		}
		return helperOnly, full.helper, nil
	}
	return nil, nil, fmt.Errorf("function %q not found", name)
}
