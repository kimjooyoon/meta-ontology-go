package bidir

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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

func TestLowerContextReplayIsDeterministicAndNonMutating(t *testing.T) {
	file := lowerContextFixture(t)
	before, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	first, err := LowerContext(context.Background(), file)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LowerContext(context.Background(), file)
	if err != nil {
		t.Fatal(err)
	}
	if first.SemanticCanonical() != second.SemanticCanonical() {
		t.Fatalf("replayed lowering is not deterministic:\n%s\n%s", first.SemanticCanonical(), second.SemanticCanonical())
	}
	after, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("successful lowering mutated syntax input")
	}
}

func assertLowerCanceled(t *testing.T, got semantic.IR, err error) {
	t.Helper()
	if !errors.Is(err, ErrLowerCanceled) || err.Error() != ErrLowerCanceled.Error() {
		t.Fatalf("cancellation error = %v", err)
	}
	if !reflect.DeepEqual(got, semantic.IR{}) {
		t.Fatalf("canceled lowerer exposed partial IR: %#v", got)
	}
}

func lowerContextFixture(t *testing.T) *syntax.File {
	t.Helper()
	file, diagnostics := syntax.ParseFile("cancel.gooo", `package billing
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder(Order) -> Payment`)
	if diagnostics.Error() != nil {
		t.Fatal(diagnostics.Error())
	}
	return file
}

func repeatedLowerDocument(count int) Document {
	document := Document{Package: "billing", Namespace: "billing"}
	for index := 0; index < count; index++ {
		document.Declarations = append(document.Declarations, Declaration{
			Kind: EntityKind,
			ID:   ID("billing://entity/item-" + zeroPad(index)),
			Name: "Item" + zeroPad(index),
		})
	}
	return document
}

func cloneLowerDocument(document Document) Document {
	clone := document
	clone.Declarations = append([]Declaration(nil), document.Declarations...)
	for index := range clone.Declarations {
		clone.Declarations[index].Inputs = append([]Reference(nil), document.Declarations[index].Inputs...)
		clone.Declarations[index].Outputs = append([]Reference(nil), document.Declarations[index].Outputs...)
	}
	clone.Relations = append([]Relation(nil), document.Relations...)
	return clone
}

func zeroPad(value int) string {
	if value < 10 {
		return "00" + string(rune('0'+value))
	}
	if value < 100 {
		return "0" + string(rune('0'+value/10)) + string(rune('0'+value%10))
	}
	return string(rune('0'+value/100)) + string(rune('0'+(value/10)%10)) + string(rune('0'+value%10))
}

type cancelAfterContext struct {
	done   chan struct{}
	limit  int32
	checks atomic.Int32
	once   sync.Once
}

func newCancelAfterContext(limit int) *cancelAfterContext {
	return &cancelAfterContext{done: make(chan struct{}), limit: int32(limit)}
}

func (c *cancelAfterContext) Deadline() (time.Time, bool) { return time.Time{}, false }

func (c *cancelAfterContext) Done() <-chan struct{} {
	if c.checks.Add(1) >= c.limit {
		c.once.Do(func() { close(c.done) })
	}
	return c.done
}

func (c *cancelAfterContext) Err() error {
	select {
	case <-c.done:
		return context.Canceled
	default:
		return nil
	}
}

func (c *cancelAfterContext) Value(any) any { return nil }
