package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestSourceSearchRealizationRetainsOtherDeclarations(t *testing.T) {
	source, prior := sourceSearchReplayFixture(t)
	realized, err := RealizeSourceAssembly(context.Background(), "search.gooo", source, prior)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(realized.Source, "assembling") || strings.Contains(realized.Source, "__GOOO_BODY_HOLE_") ||
		realized.FinitePassed != 5 || realized.FiniteTotal != 5 || realized.ModelCalls != 0 || realized.OriginalSourceSHA256 != digest(source) {
		t.Fatal("source search checkpoint evidence differs", realized)
	}
	replayed, err := Generate("search.gooo", []byte(realized.Source), prior.Report.Activity)
	if err != nil || replayed.Source != prior.Source {
		t.Fatal("checkpoint does not reproduce selected native projection", err)
	}
	if !strings.Contains(realized.Source, "entity Integer id") {
		t.Fatal("realization removed an unrelated declaration")
	}
}

func TestSourceSearchPreflightChecksGrammarAndOrdinaryReferences(t *testing.T) {
	source := []byte(`package offset
namespace offset
entity Integer id "offset://integer"
activity Add(Integer) -> Integer computes "return input + __GOOO_BODY_HOLE_offset__" assembling {
    search hole "offset" grammar "integer-offset-constant/v1" intent "Choose a typed offset." max_candidates "16"
    case "2" -> "3"
    attempts "8"
}`)
	if err := ValidateSourceAssembly(context.Background(), "offset.gooo", source, "Add"); err != nil {
		t.Fatal("a valid source search was rejected", err)
	}
	bad := []byte(strings.Replace(string(source), "return input +", "return missing +", 1))
	if err := ValidateSourceAssembly(context.Background(), "offset.gooo", bad, "Add"); err == nil {
		t.Fatal("an undeclared ordinary reference passed preflight")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := ValidateSourceAssembly(ctx, "offset.gooo", source, "Add"); err == nil {
		t.Fatal("canceled preflight succeeded")
	}
}
