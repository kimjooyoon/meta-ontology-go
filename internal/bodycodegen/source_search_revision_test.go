package bodycodegen

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestSearchRevisionOnlyChangesSelectedGrammarBoundAndBudget(t *testing.T) {
	source, err := os.ReadFile("../../examples/search-feedback/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	source = append(source, []byte("\nactivity Echo(Integer) -> Integer computes \"return input\"\n")...)
	ctx := context.Background()
	before, err := SourceAssembly(ctx, "search.gooo", source, "Add")
	if err != nil {
		t.Fatal(err)
	}
	next, err := ReviseAssemblySearch(ctx, "search.gooo", source, "Add", "contextual", 8)
	if err != nil {
		t.Fatal(err)
	}
	after, err := SourceAssembly(ctx, "search.gooo", next, "Add")
	before.Search.Grammar, before.Search.MaxCandidates, before.MaxAttempts = "integer-hole-residual/v1", 16, 8
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("search revision changed another source obligation", err)
	}
	start, end := bytes.Index(source, []byte("assembling")), bytes.Index(source, []byte("\nactivity Echo"))
	if !bytes.HasPrefix(next, source[:start]) || !bytes.HasSuffix(next, source[end:]) {
		t.Fatal("search revision changed the body, identity or another declaration")
	}
	for _, tc := range []struct {
		id       string
		attempts int
	}{{"missing", 1}, {"wider", 17}, {"wider", 0}} {
		if _, err = ReviseAssemblySearch(ctx, "search.gooo", source, "Add", tc.id, tc.attempts); err == nil {
			t.Fatal("undeclared or out-of-bound search change accepted", tc)
		}
	}
	if _, err = ReviseAssemblySearch(ctx, "search.gooo", next, "Add", "contextual", 8); err == nil || !strings.Contains(err.Error(), "must change") {
		t.Fatal("same grammar/bound was repeated", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = ReviseAssemblySearch(canceled, "search.gooo", source, "Add", "wider", 8); err == nil {
		t.Fatal("canceled revision changed source")
	}
}
