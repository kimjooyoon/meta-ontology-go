package bidir

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReconcileRejectsInvalidEvidenceTransactionally(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	bad := NewSourcedFact(
		DeterministicFact,
		"billing://activity/pay-order",
		PredicateInvokes,
		"billing://activity/audit-payment",
		SourceSpan{File: "payment.go", Start: 20, End: 10},
	)

	result, err := Reconcile(base, FactDelta{Added: FactSet{bad}})
	if err == nil {
		t.Fatal("invalid source span was accepted")
	}
	var reconcileErr *ReconcileError
	if !errors.As(err, &reconcileErr) || len(reconcileErr.Conflicts) != 1 || reconcileErr.Conflicts[0].Kind != ConflictInvalidFact {
		t.Fatalf("unexpected invalid evidence error: %v", err)
	}
	if !SemanticEquivalent(base, result.Model) || !result.Delta.IsEmpty() {
		t.Fatalf("invalid evidence changed the model: result=%#v", result)
	}
}
func assertPutNoWrite(t *testing.T, err error, code string) {
	t.Helper()
	var putErr *PutError
	if !errors.As(err, &putErr) || !putErr.NoWrite || putErr.Code != code {
		t.Fatalf("unexpected Put rejection: %v", err)
	}
}
func billingGoooDocument(t *testing.T) Document {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locating dogfood test failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", ".."))
	path := filepath.Join(root, "examples", "billing", "main.gooo")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	file, diagnostics := syntax.ParseFile(path, string(source))
	if diagnostics.Error() != nil {
		t.Fatalf("parse %s: %v", path, diagnostics.Error())
	}
	if _, err := Lower(file); err != nil {
		t.Fatalf("lower %s to semantic IR: %v", path, err)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatalf("adapt %s: %v", path, err)
	}
	return document
}
