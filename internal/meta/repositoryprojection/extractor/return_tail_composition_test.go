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

func TestReturnTailCalleeProofComposition(t *testing.T) {
	source := "package p\n\nfunc caller() error {\n\thelper()\n\thelper()\n\treturn nil\n}\n\nfunc helper() error { return leaf() }\n\nfunc leaf() error { return nil }\n"
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
	var caller *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "caller" {
			caller = function
			break
		}
	}
	if caller == nil {
		t.Fatal("composition fixture lacks caller")
	}
	evidence, err := checkTypes(root, "x.go", fset, file, caller)
	if err != nil {
		t.Fatal(err)
	}
	evidence.contractSourceDigest = "contract-source"
	evidence.contractSemanticDigest = "contract-semantic"
	declarations := make(map[string]*ast.FuncDecl)
	objects := make(map[string]*types.Func)
	for object, declaration := range evidence.funcs {
		declarations[object.Name()] = declaration
		objects[object.Name()] = object
	}
	leafProof := returnTailTestHelperProof(t, "leaf", evidence, objects["leaf"], declarations["leaf"], "leaf-evidence", nil)
	registry := map[string]returnTailHelperProof{"leaf": leafProof}
	helperCall, ok := declarations["helper"].Body.List[0].(*ast.ReturnStmt)
	if !ok || len(helperCall.Results) != 1 {
		t.Fatal("composition fixture lacks helper return call")
	}
	leafDependency, ok := returnTailCalleeDependency(helperCall.Results[0].(*ast.CallExpr), evidence, registry)
	if !ok {
		t.Fatal("registered leaf proof did not produce a dependency")
	}
	helperProof := returnTailTestHelperProof(t, "helper", evidence, objects["helper"], declarations["helper"], "helper-evidence", []CalleeDependencyEvidence{leafDependency})
	registry["helper"] = helperProof
	context := &returnTailValidationContext{
		visiting:        make(map[*types.Func]bool),
		memo:            make(map[*types.Func]returnTailValidation),
		proofBodyVisits: make(map[*types.Func]int),
	}
	dependencies, err := returnTailCalleeEffectsWithContext(caller.Body.List, evidence, registry, context)
	if err != nil {
		t.Fatal(err)
	}
	if len(dependencies) != 2 || dependencies[0].Name != "helper" || dependencies[1].Name != "helper" {
		t.Fatalf("composed dependencies=%+v, want two registry-backed helper dependencies", dependencies)
	}
	if context.proofBodyVisits[objects["helper"]] != 1 || context.proofBodyVisits[objects["leaf"]] != 1 {
		t.Fatalf("proof body visits=%+v, want helper=1 leaf=1", context.proofBodyVisits)
	}
}

func TestReturnTailCalleeProofCycleFailsClosed(t *testing.T) {
	source := "package p\n\nfunc caller() error { return cycleA() }\n\nfunc cycleA() error { return cycleB() }\n\nfunc cycleB() error { return cycleA() }\n"
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
	caller, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		t.Fatal("cycle fixture lacks caller")
	}
	evidence, err := checkTypes(root, "x.go", fset, file, caller)
	if err != nil {
		t.Fatal(err)
	}
	_, err = returnTailCalleeEffects(caller.Body.List, evidence, nil)
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "CALLEE_EFFECTS_UNPROVEN" || failure.UnknownClass != "DIRECT_MISSING" {
		t.Fatalf("cycle error=%v, want fail-closed callee-effects failure", err)
	}
}

func returnTailTestHelperProof(t *testing.T, name string, evidence typeEvidence, object *types.Func, declaration *ast.FuncDecl, evidenceID string, dependencies []CalleeDependencyEvidence) returnTailHelperProof {
	t.Helper()
	signatureDigest, bodyDigest, ok := returnTailFunctionDigests(evidence.fset, declaration)
	if !ok || object == nil || declaration == nil {
		t.Fatalf("helper proof inputs invalid: object=%v declaration=%v", object, declaration)
	}
	return returnTailHelperProof{
		helperName:              name,
		helperObjectIdentity:    returnTailObjectIdentity(object, evidence),
		helperSignatureDigest:   signatureDigest,
		helperBodyDigest:        bodyDigest,
		helperType:              returnTailCanonicalType(object.Type()),
		contractSourceDigest:    evidence.contractSourceDigest,
		contractSemanticDigest:  evidence.contractSemanticDigest,
		calleeEffectsEvidenceID: evidenceID,
		dependencies:            dependencies,
	}
}
