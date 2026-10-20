package extractor

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func freshStandardRoot31711(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module experiment.local/origin31711\n\ngo 1.27.3\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestStandardTypeStageOrigins31711(t *testing.T) {
	root := freshStandardRoot31711(t)
	session := newStandardImportSession()
	first := newModuleImporterWithStandard(root, session)
	start := time.Now()
	firstContext, err := first.Import("context")
	if err != nil {
		t.Fatal(err)
	}
	firstElapsed := time.Since(start)
	second := newModuleImporterWithStandard(root, session)
	start = time.Now()
	secondContext, err := second.Import("context")
	if err != nil {
		t.Fatal(err)
	}
	secondElapsed := time.Since(start)
	if firstContext != secondContext || first.listCalls == 0 || second.listCalls != 0 || session.current != nil {
		t.Fatalf("standard origin or actual list work changed: calls=%d/%d", first.listCalls, second.listCalls)
	}
	if err := os.Mkdir(filepath.Join(root, "inner"), 0700); err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(root, "inner", "value.go")
	if err := os.WriteFile(local, []byte("package inner\ntype Count int\nconst Value Count = 37\n"), 0600); err != nil {
		t.Fatal(err)
	}
	before := freshStandardEvidence31711(t, root, session)
	if err := os.WriteFile(local, []byte("package inner\ntype Count float64\nconst Value Count = 37.5\n"), 0600); err != nil {
		t.Fatal(err)
	}
	after := freshStandardEvidence31711(t, root, session)
	if before.info == after.info || before.pkg == after.pkg || freshResultKind31711(t, before) != types.Int || freshResultKind31711(t, after) != types.Float64 {
		t.Fatal("project source was reused after its original bytes changed")
	}
	if freshImportedPackage31711(t, before, "context") != freshImportedPackage31711(t, after, "context") ||
		freshImportedPackage31711(t, before, "experiment.local/origin31711/inner") == freshImportedPackage31711(t, after, "experiment.local/origin31711/inner") {
		t.Fatal("standard and project origins were combined")
	}
	if out := os.Getenv("GOOO_FRESH_SOURCE_OUTPUT"); out != "" {
		raw, err := json.MarshalIndent(map[string]any{"scope": "one synthetic source-owned session", "first_go_list_calls": first.listCalls, "second_go_list_calls": second.listCalls,
			"first_elapsed_ns": firstElapsed.Nanoseconds(), "second_elapsed_ns": secondElapsed.Nanoseconds(), "standard_type_identity_preserved": true,
			"project_source_before": "int", "project_source_after": "float64", "fresh_project_types": true, "historical_cohort_reruns": 0, "model_calls": 0}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "standard31711.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func freshStandardEvidence31711(t *testing.T, root string, session *standardImportSession) typeEvidence {
	t.Helper()
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "main.go", `package origin31711
import "context"
import "experiment.local/origin31711/inner"
func F(ctx context.Context) inner.Count { _ = ctx; return inner.Value }
`, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var function *ast.FuncDecl
	for _, declaration := range file.Decls {
		if found, ok := declaration.(*ast.FuncDecl); ok {
			function = found
		}
	}
	evidence, err := checkTypesWithStandard(root, "main.go", set, file, function, session)
	if err != nil {
		t.Fatal(err)
	}
	return evidence
}

func freshResultKind31711(t *testing.T, evidence typeEvidence) types.BasicKind {
	t.Helper()
	function := evidence.pkg.Scope().Lookup("F").(*types.Func)
	return function.Type().(*types.Signature).Results().At(0).Type().Underlying().(*types.Basic).Kind()
}

func freshImportedPackage31711(t *testing.T, evidence typeEvidence, name string) *types.Package {
	t.Helper()
	for _, imported := range evidence.pkg.Imports() {
		if imported.Path() == name {
			return imported
		}
	}
	t.Fatalf("missing import %s", name)
	return nil
}

func TestStandardImportScopeChanges31711(t *testing.T) {
	root := freshStandardRoot31711(t)
	session := newStandardImportSession()
	first := newModuleImporterWithStandard(root, session)
	previous, err := first.Import("context")
	if err != nil {
		t.Fatal(err)
	}
	module := filepath.Join(root, "go.mod")
	if err := os.WriteFile(module, []byte("module experiment.local/origin31711\n\ngo 1.27.3\n// a new module origin\n"), 0600); err != nil {
		t.Fatal(err)
	}
	changed := newModuleImporterWithStandard(root, session)
	next, err := changed.Import("context")
	if err != nil {
		t.Fatal(err)
	}
	if next == previous || changed.listCalls == 0 {
		t.Fatal("changed module reused the prior scope")
	}
	previous = next
	t.Setenv("GOOO_STANDARD_IMPORT_FRESH_SCOPE", "31711")
	changed = newModuleImporterWithStandard(root, session)
	next, err = changed.Import("context")
	if err != nil {
		t.Fatal(err)
	}
	if next == previous || changed.listCalls == 0 {
		t.Fatal("changed execution environment reused the prior scope")
	}
	other := newModuleImporterWithStandard(freshStandardRoot31711(t), session)
	otherPackage, err := other.Import("context")
	if err != nil {
		t.Fatal(err)
	}
	if otherPackage == next || other.listCalls == 0 {
		t.Fatal("another project root reused the prior scope")
	}
	if err := os.Remove(module); err != nil {
		t.Fatal(err)
	}
	missing := newModuleImporterWithStandard(root, session)
	if _, err := missing.Import("context"); err == nil {
		t.Fatal("missing source scope returned cached standard types")
	}
}

func TestStandardConcurrentStages31711(t *testing.T) {
	root := freshStandardRoot31711(t)
	session := newStandardImportSession()
	const count = 6
	stages := make([]*moduleImporter, count)
	packages := make([]*types.Package, count)
	errors := make([]error, count)
	var done sync.WaitGroup
	for index := range count {
		stages[index] = newModuleImporterWithStandard(root, session)
		done.Go(func() { packages[index], errors[index] = stages[index].Import("context") })
	}
	wait := make(chan struct{})
	go func() { done.Wait(); close(wait) }()
	select {
	case <-wait:
	case <-time.After(45 * time.Second):
		t.Fatal("same-session standard imports deadlocked")
	}
	calls := 0
	for index := range count {
		if errors[index] != nil || packages[index] == nil || packages[index] != packages[0] {
			t.Fatalf("stage %d lost source identity: %v", index, errors[index])
		}
		calls += stages[index].listCalls
	}
	if calls != 1 || session.current != nil {
		t.Fatalf("parallel stages repeated list work: %d", calls)
	}
}
