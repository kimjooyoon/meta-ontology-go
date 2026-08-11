package fuzz

import (
	"reflect"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	maxSourceBytes = 8 << 10
	operationLimit = 250 * time.Millisecond
	fuzzFilename   = "fuzz.gooo"
)

type lexResult struct {
	Tokens      syntax.Tokens
	Diagnostics syntax.Diagnostics
}

type parseResult struct {
	File        *syntax.File
	Diagnostics syntax.Diagnostics
}

type timedResult[T any] struct {
	value      T
	panicValue any
}

func FuzzLexConformance(f *testing.F) {
	addMinimalSeeds(f)
	f.Fuzz(func(t *testing.T, source string) {
		limitSource(t, source)
		first := runWithLimit(t, "lex", func() lexResult {
			tokens, diagnostics := syntax.LexFile(fuzzFilename, source)
			return lexResult{Tokens: tokens, Diagnostics: diagnostics}
		})
		second := runWithLimit(t, "lex", func() lexResult {
			tokens, diagnostics := syntax.LexFile(fuzzFilename, source)
			return lexResult{Tokens: tokens, Diagnostics: diagnostics}
		})
		if !reflect.DeepEqual(first, second) {
			t.Fatalf("lexing was not deterministic:\nfirst=%#v\nsecond=%#v", first, second)
		}
		assertLexResult(t, source, first)
	})
}

func FuzzParseConformance(f *testing.F) {
	addMinimalSeeds(f)
	f.Fuzz(func(t *testing.T, source string) {
		limitSource(t, source)
		first := runWithLimit(t, "parse", func() parseResult {
			file, diagnostics := syntax.ParseFile(fuzzFilename, source)
			return parseResult{File: file, Diagnostics: diagnostics}
		})
		second := runWithLimit(t, "parse", func() parseResult {
			file, diagnostics := syntax.ParseFile(fuzzFilename, source)
			return parseResult{File: file, Diagnostics: diagnostics}
		})
		if !reflect.DeepEqual(first, second) {
			t.Fatalf("parsing was not deterministic:\nfirst=%#v\nsecond=%#v", first, second)
		}
		assertDiagnostics(t, source, first.Diagnostics)
		assertFile(t, source, first.File)
	})
}

func limitSource(t *testing.T, source string) {
	t.Helper()
	if len(source) > maxSourceBytes {
		t.Skipf("input is %d bytes; fuzz limit is %d", len(source), maxSourceBytes)
	}
}

func runWithLimit[T any](t *testing.T, label string, operation func() T) T {
	t.Helper()
	done := make(chan timedResult[T], 1)
	go func() {
		result := timedResult[T]{}
		defer func() {
			result.panicValue = recover()
			done <- result
		}()
		result.value = operation()
	}()
	select {
	case result := <-done:
		if result.panicValue != nil {
			t.Fatalf("%s panicked: %v", label, result.panicValue)
		}
		return result.value
	case <-time.After(operationLimit):
		t.Fatalf("%s exceeded %s", label, operationLimit)
	}
	var zero T
	return zero
}
