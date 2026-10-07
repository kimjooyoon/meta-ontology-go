package bodycodegen

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"testing"
)

func TestReviseAssemblyBudgetPreservesContractAndSurroundings(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/record-candidate-continuation.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	before, err := SourceAssembly(ctx, "budget.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	same, err := ReviseAssemblyBudget(ctx, "budget.gooo", source, "Select", before.MaxAttempts)
	if err != nil || !bytes.Equal(same, source) {
		t.Fatal("unchanged budget rewrote the source", err)
	}
	next, err := ReviseAssemblyBudget(ctx, "budget.gooo", source, "Select", 2)
	if err != nil {
		t.Fatal(err)
	}
	after, err := SourceAssembly(ctx, "budget.gooo", next, "Select")
	before.MaxAttempts = 2
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("budget revision changed another contract field", err)
	}
	start, end := bytes.Index(source, []byte("assembling")), bytes.Index(source, []byte("\nactivity Label"))
	if !bytes.HasPrefix(next, source[:start]) || !bytes.HasSuffix(next, source[end:]) {
		t.Fatal("budget revision modified another declaration or the computes body")
	}
	for _, count := range []int{0, 65} {
		if _, err := ReviseAssemblyBudget(ctx, "budget.gooo", source, "Select", count); err == nil {
			t.Fatal("invalid attempt count accepted")
		}
	}
	if _, err := ReviseAssemblyBudget(ctx, "budget.gooo", source, "Label", 2); err == nil {
		t.Fatal("ordinary activity acquired an assembly")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := ReviseAssemblyBudget(canceled, "budget.gooo", source, "Select", 2); err == nil {
		t.Fatal("canceled source rewrite succeeded")
	}
}
