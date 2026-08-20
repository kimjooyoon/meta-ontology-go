package bidir

import (
	"context"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
	"time"
)

func TestLowerContextCancellationIsStableAndNonMutating(t *testing.T) {
	file := lowerContextFixture(t)
	before, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := LowerContext(ctx, file)
	assertLowerCanceled(t, got, err)
	after, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("canceled lowering mutated syntax input:\nbefore %#v\nafter %#v", before, after)
	}
}
func TestLowerContextDeadlineUsesStableCancellation(t *testing.T) {
	file := lowerContextFixture(t)
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	got, err := LowerContext(ctx, file)
	assertLowerCanceled(t, got, err)
}
func TestLowerDocumentContextCancelsDuringDeterministicLoop(t *testing.T) {
	document := repeatedLowerDocument(64)
	before := cloneLowerDocument(document)
	ctx := newCancelAfterContext(5)
	started := time.Now()
	got, err := LowerDocumentContext(ctx, document)
	if !errors.Is(err, ErrLowerCanceled) {
		t.Fatalf("mid-loop cancellation error = %v", err)
	}
	if !reflect.DeepEqual(got, semantic.IR{}) {
		t.Fatalf("canceled lowerer exposed partial IR: %#v", got)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("cancellation was not prompt: %s", elapsed)
	}
	if !reflect.DeepEqual(document, before) {
		t.Fatal("canceled document lowering mutated its input")
	}
}
