package extractor

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestRenderedCapacityObservationAcceptanceSet(t *testing.T) {
	cases := []struct {
		name string
		source string
		check func(t *testing.T, selection *renderedCapacitySelection, err error)
		render func(*token.FileSet, *ast.File, []byte, *ast.FuncDecl) ([]byte, error)
	}{
		{
			name:   "01-measured-within-both-caps",
			source: "package p\n\nfunc F() error {\n\treturn nil\n}\n",
			check: func(t *testing.T, selection *renderedCapacitySelection, err error) {
				t.Helper()
				if err != nil || selection == nil || selection.function != nil || len(selection.observations) != 1 {
					t.Fatalf("selection=%+v err=%v, want measured no-selection", selection, err)
				}
				observation := selection.observations[0]
				if observation.functionStatus != renderedCapacityWithinCap || observation.helperStatus != renderedCapacityWithinCap || observation.helperLines == nil {
					t.Fatalf("observation=%+v, want both measurements within cap", observation)
				}
			},
		},
		{
			name:   "02-known-function-span-overflow",
			source: "package p\n\nfunc F() error {\n" + strings.Repeat("\t_ = 1\n", functionLineLimit+1) + "\treturn nil\n}\n",
			check: func(t *testing.T, selection *renderedCapacitySelection, err error) {
				t.Helper()
				if err != nil || selection == nil || selection.function == nil || selection.function.Name.Name != "F" || len(selection.observations) != 1 {
					t.Fatalf("selection=%+v err=%v, want F selected", selection, err)
				}
				observation := selection.observations[0]
				if observation.functionStatus != renderedCapacityOverCap || observation.functionLines <= functionLineLimit || observation.sourceDigest == "" {
					t.Fatalf("observation=%+v, want measured function-span violation", observation)
				}
			},
		},
		{
			name:   "03-short-declaration-oversized-documentation-artifact",
			source: "package p\n\n" + strings.Repeat("// documentation\n", functionLineLimit+1) + "func F() error { return nil }\n",
			check: func(t *testing.T, selection *renderedCapacitySelection, err error) {
				t.Helper()
				if err != nil || selection == nil || selection.function == nil || selection.function.Name.Name != "F" {
					t.Fatalf("selection=%+v err=%v, want F selected", selection, err)
				}
				observation := selection.observations[0]
				if observation.functionStatus != renderedCapacityWithinCap || observation.helperStatus != renderedCapacityOverCap || observation.helperLines == nil || *observation.helperLines <= functionLineLimit {
					t.Fatalf("observation=%+v, want oversized rendered documentation artifact", observation)
				}
			},
		},
		{
			name:   "04-failed-helper-measurement-propagates-structured-error",
			source: "package p\n\nfunc F() error {\n\treturn nil\n}\n",
			check: func(t *testing.T, selection *renderedCapacitySelection, err error) {
				t.Helper()
				var failure Failure
				if selection != nil || !errors.As(err, &failure) || failure.Reason != "PREFLIGHT_RENDER_FAILED" {
					t.Fatalf("selection=%+v err=%v, want structured helper measurement failure", selection, err)
				}
				if len(failure.Diagnostics) == 0 || !strings.Contains(failure.Diagnostics[0], "measurement=UNMEASURED") {
					t.Fatalf("failure=%+v, want unresolved measurement diagnostics", failure)
				}
			},
			render: func(_ *token.FileSet, _ *ast.File, _ []byte, _ *ast.FuncDecl) ([]byte, error) {
				return nil, fail("observe-plan", "render-capacity", "PREFLIGHT_RENDER_FAILED", "DIRECT_MISSING", "restore-render-evidence", nil)
			},
		},
		{
			name:   "05-earlier-unmeasured-later-known-violation",
			source: "package p\n\nfunc Earlier() error { return nil }\n\nfunc Later() error {\n" + strings.Repeat("\t_ = 1\n", functionLineLimit+1) + "\treturn nil\n}\n",
			check: func(t *testing.T, selection *renderedCapacitySelection, err error) {
				t.Helper()
				if err != nil || selection == nil || selection.function == nil || selection.function.Name.Name != "Later" || len(selection.observations) != 2 {
					t.Fatalf("selection=%+v err=%v, want later known violation selected", selection, err)
				}
				earlier, later := selection.observations[0], selection.observations[1]
				if earlier.helperStatus != renderedCapacityUnmeasured || earlier.helperLines != nil || earlier.helperFailure == nil || later.functionStatus != renderedCapacityOverCap {
					t.Fatalf("observations=%+v, want unresolved earlier and known later violation", selection.observations)
				}
			},
			render: func(fset *token.FileSet, file *ast.File, source []byte, function *ast.FuncDecl) ([]byte, error) {
				if function.Name.Name == "Earlier" {
					return nil, fail("observe-plan", "render-capacity", "PREFLIGHT_RENDER_FAILED", "DIRECT_MISSING", "restore-render-evidence", nil)
				}
				return renderedDeclarationHelper(fset, file, source, function)
			},
		},
		{
			name:   "06-same-named-method-and-free-function",
			source: "package p\n\ntype T struct{}\n\nfunc (T) F() error { return nil }\n\nfunc F() error {\n" + strings.Repeat("\t_ = 1\n", functionLineLimit+1) + "\treturn nil\n}\n",
			check: func(t *testing.T, selection *renderedCapacitySelection, err error) {
				t.Helper()
				if err != nil || selection == nil || selection.function == nil || selection.function.Recv != nil || selection.function.Name.Name != "F" {
					t.Fatalf("selection=%+v err=%v, want free function F selected", selection, err)
				}
				if len(selection.observations) != 2 || !strings.HasPrefix(selection.observations[0].subject, "method:") || selection.observations[0].receiver != "T" || selection.observations[1].subject != "func:F" {
					t.Fatalf("observations=%+v, want method/free identities", selection.observations)
				}
				if selection.observations[0].functionStart == selection.observations[1].functionStart || selection.observations[0].declarationStart == selection.observations[1].declarationStart {
					t.Fatalf("observations=%+v, want distinct source coordinates", selection.observations)
				}
			},
		},
	}

	if len(cases) != 6 {
		t.Fatalf("rendered capacity acceptance denominator=%d, want 6", len(cases))
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "fixture.go", testCase.source, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			renderer := renderedDeclarationHelper
			if testCase.render != nil {
				renderer = testCase.render
			}
			selection, selectionErr := firstOversizedFunctionWithRenderer(fset, file, []byte(testCase.source), renderer)
			testCase.check(t, selection, selectionErr)
		})
	}
}
