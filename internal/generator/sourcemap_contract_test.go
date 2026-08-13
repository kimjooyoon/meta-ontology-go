package generator

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestSourceMapRepeatAndPermutationAreStable(t *testing.T) {
	firstIR := acceptanceFixture()
	secondIR := acceptanceFixture()
	secondIR.Entities[0], secondIR.Entities[1] = secondIR.Entities[1], secondIR.Entities[0]
	first, err := Generate(firstIR, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Generate(secondIR, nil)
	if err != nil {
		t.Fatal(err)
	}
	repeat, err := Generate(firstIR, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Source, second.Source) || !reflect.DeepEqual(first.SourceMap, second.SourceMap) {
		t.Fatal("declaration permutation changed source-map identity")
	}
	if !reflect.DeepEqual(first.SourceMap, repeat.SourceMap) {
		t.Fatal("repeated source-map generation changed identity")
	}
}

func TestSourceMapRejectsStaleSlotIdentity(t *testing.T) {
	ir := acceptanceFixture()
	result := mustAcceptanceResult(t, ir, nil)
	ir.Activities[0].Slots[0].ID = "gooo://slot/new-implementation"
	stale := string(result.Source)
	if _, err := Generate(ir, []byte(stale)); err == nil || !strings.Contains(err.Error(), "stale slot identity") {
		t.Fatalf("expected stale slot rejection, got %v", err)
	}
}

func TestSourceMapFailureDoesNotMutateIROrSource(t *testing.T) {
	ir := acceptanceFixture()
	result := mustAcceptanceResult(t, ir, nil)
	corrupt := append([]byte(nil), result.Source...)
	corrupt = append(corrupt, []byte("\n//gooo:slot:start id=\"orphan\"\n")...)
	beforeIR := copyIR(ir)
	beforeSource := append([]byte(nil), corrupt...)
	if _, err := Generate(ir, corrupt); err == nil {
		t.Fatal("orphan source-map input was accepted")
	}
	if !reflect.DeepEqual(ir, beforeIR) || !bytes.Equal(corrupt, beforeSource) {
		t.Fatal("source-map failure mutated caller-owned input")
	}
}
