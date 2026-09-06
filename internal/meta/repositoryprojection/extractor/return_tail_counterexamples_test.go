package extractor

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"testing"
)

type returnTailCalleeCounterexample struct {
	name          string
	source        string
	proofMutation func(*returnTailHelperProof)
	assertSafe    func(*testing.T, ast.Node, typeEvidence)
	noProof       bool
}

func TestReturnTailCalleeProofCounterexamples(t *testing.T) {
	cases := []returnTailCalleeCounterexample{
		{
			name:    "no proof for generated-looking helper",
			noProof: true,
			source:  "package p\n\nfunc caller() error { return callerExtractedReturnTail99() }\n\nfunc callerExtractedReturnTail99() error { return nil }\n",
		},
		{
			name:   "body digest drift",
			source: returnTailProofCounterexampleBaseSource(),
			proofMutation: func(proof *returnTailHelperProof) {
				proof.helperBodyDigest = "body-drift"
			},
		},
		{
			name:   "signature digest drift",
			source: returnTailProofCounterexampleBaseSource(),
			proofMutation: func(proof *returnTailHelperProof) {
				proof.helperSignatureDigest = "signature-drift"
			},
		},
		{
			name:   "object identity drift",
			source: returnTailProofCounterexampleBaseSource(),
			proofMutation: func(proof *returnTailHelperProof) {
				proof.helperObjectIdentity = "object-drift"
			},
		},
		{
			name:   "contract digest drift",
			source: returnTailProofCounterexampleBaseSource(),
			proofMutation: func(proof *returnTailHelperProof) {
				proof.contractSourceDigest = "contract-drift"
			},
		},
		{
			name:   "dependency evidence drift",
			source: returnTailProofCounterexampleBaseSource(),
			proofMutation: func(proof *returnTailHelperProof) {
				proof.dependencies = []CalleeDependencyEvidence{{Name: "unexpected", EvidenceID: "unexpected"}}
			},
		},
		{
			name:   "global write",
			source: "package p\n\nvar counter int\n\nfunc caller() error { return helper() }\n\nfunc helper() error { counter = 1; return nil }\n",
			assertSafe: func(t *testing.T, node ast.Node, evidence typeEvidence) {
				helper, ok := node.(*ast.FuncDecl)
				if !ok {
					t.Fatal("global-write fixture lacks helper declaration")
				}
				_, _, safe := returnTailGlobalReadEvidence(helper, evidence)
				if safe {
					t.Fatal("global write was reported safe")
				}
			},
		},
		{
			name:   "global address escape",
			source: "package p\n\nvar target int\n\nfunc caller() error { return helper() }\n\nfunc helper() error { _ = &target; return nil }\n",
			assertSafe: func(t *testing.T, node ast.Node, evidence typeEvidence) {
				helper, ok := node.(*ast.FuncDecl)
				if !ok {
					t.Fatal("global-address fixture lacks helper declaration")
				}
				_, _, safe := returnTailGlobalReadEvidence(helper, evidence)
				if safe {
					t.Fatal("global address escape was reported safe")
				}
			},
		},
		{
			name:   "closure capture",
			source: "package p\n\nfunc caller() error { return helper() }\n\nfunc helper() error { value := 0; set := func() { value = 1 }; _ = set; return nil }\n",
			assertSafe: func(t *testing.T, node ast.Node, evidence typeEvidence) {
				helper, ok := node.(*ast.FuncDecl)
				if !ok {
					t.Fatal("closure fixture lacks helper declaration")
				}
				_, _, safe := returnTailGlobalReadEvidence(helper, evidence)
				if safe {
					t.Fatal("closure capture was reported safe")
				}
			},
		},
		{
			name:   "parenthesized global channel receive",
			source: "package p\n\nvar globalChannel chan int\n\nfunc consume(int) {}\n\nfunc caller() error { return helper() }\n\nfunc helper() error { consume((<-globalChannel)); return nil }\n",
			assertSafe: func(t *testing.T, node ast.Node, evidence typeEvidence) {
				helper, ok := node.(*ast.FuncDecl)
				if !ok {
					t.Fatal("channel-receive fixture lacks helper declaration")
				}
				_, _, safe := returnTailGlobalReadEvidence(helper, evidence)
				if safe {
					t.Fatal("global channel receive was reported safe")
				}
			},
		},
		{
			name:   "parenthesized global channel range",
			source: "package p\n\nvar globalChannel chan int\n\nfunc caller() error { return helper() }\n\nfunc helper() error { for range (globalChannel) {}; return nil }\n",
			assertSafe: func(t *testing.T, node ast.Node, evidence typeEvidence) {
				helper, ok := node.(*ast.FuncDecl)
				if !ok {
					t.Fatal("channel-range fixture lacks helper declaration")
				}
				_, _, safe := returnTailGlobalReadEvidence(helper, evidence)
				if safe {
					t.Fatal("global channel range was reported safe")
				}
			},
		},
		{
			name:   "derived channel range",
			source: "package p\n\nfunc caller() error { return helper(nil) }\n\nfunc helper(channel chan int) error { derived := channel; for range (derived) {}; return nil }\n",
		},
		{
			name:   "derived channel receive",
			source: "package p\n\nfunc consume(int) {}\n\nfunc caller() error { return helper(nil) }\n\nfunc helper(channel chan int) error { derived := channel; consume((<-derived)); return nil }\n",
		},
		{
			name:   "nested call inside typed conversion",
			source: "package p\n\ntype conversionResult int\n\nfunc (conversionResult) Error() string { return \"\" }\n\nfunc unsafeConversionValue() conversionResult { panic(\"unsafe\") }\n\nfunc caller() error { return helper() }\n\nfunc helper() error { return error(unsafeConversionValue()) }\n",
		},
	}
	if len(cases) != 14 {
		t.Fatalf("callee counterexample denominator=%d, want 14", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			evidence, caller, helper := returnTailCounterexampleEvidence(t, tc.source)
			if tc.assertSafe != nil {
				tc.assertSafe(t, helper, evidence)
			}
			if tc.noProof {
				_, err := returnTailCalleeEffects(caller.Body.List, evidence, nil)
				assertReturnTailCalleeCounterexample(t, err)
				return
			}
			if helper == nil || helper.Name == nil {
				t.Fatal("callee counterexample lacks helper declaration")
			}
			helperObject, ok := evidence.info.Defs[helper.Name].(*types.Func)
			if !ok || helperObject == nil {
				t.Fatal("callee counterexample lacks typed helper")
			}
			proof := returnTailTestHelperProof(t, "helper", evidence, helperObject, helper, "helper-evidence", nil)
			if tc.assertSafe == nil && !returnTailHelperProofMatches(helperObject, helper, evidence, proof) {
				t.Fatal("baseline helper proof did not match complete typed evidence")
			}
			if tc.proofMutation != nil {
				if _, err := returnTailCalleeEffects(caller.Body.List, evidence, map[string]returnTailHelperProof{"helper": proof}); err != nil {
					t.Fatalf("complete helper proof was rejected before mutation: %v", err)
				}
				tc.proofMutation(&proof)
			}
			_, err := returnTailCalleeEffects(caller.Body.List, evidence, map[string]returnTailHelperProof{"helper": proof})
			assertReturnTailCalleeCounterexample(t, err)
		})
	}
}

func returnTailCounterexampleEvidence(t *testing.T, source string) (typeEvidence, *ast.FuncDecl, *ast.FuncDecl) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "x.go", []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var caller, helper *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name == nil {
			continue
		}
		switch function.Name.Name {
		case "caller":
			caller = function
		case "helper":
			helper = function
		}
	}
	if caller == nil {
		t.Fatal("callee counterexample lacks caller declaration")
	}
	evidence, err := checkTypes(root, "x.go", fset, file, caller)
	if err != nil {
		t.Fatal(err)
	}
	evidence.contractSourceDigest = "contract-source"
	evidence.contractSemanticDigest = "contract-semantic"
	return evidence, caller, helper
}

func assertReturnTailCalleeCounterexample(t *testing.T, err error) {
	t.Helper()
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "CALLEE_EFFECTS_UNPROVEN" || failure.UnknownClass != "DIRECT_MISSING" {
		t.Fatalf("callee counterexample error=%v, want fail-closed callee-effects failure", err)
	}
}

func returnTailProofCounterexampleBaseSource() string {
	return "package p\n\nfunc caller() error { return helper() }\n\nfunc helper() error { return nil }\n"
}
