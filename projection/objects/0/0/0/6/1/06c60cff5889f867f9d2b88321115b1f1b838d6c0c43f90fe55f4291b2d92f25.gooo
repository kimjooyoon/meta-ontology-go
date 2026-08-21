package linecaps

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestInvalidInputsFailClosed(t *testing.T) {
	root := t.TempDir()
	for name, run := range map[string]func() error{
		"invalid limits": func() error {
			_, err := Analyze(root, nil, Limits{})
			return err
		},
		"empty root": func() error {
			_, err := Analyze("", nil, DefaultLimits())
			return err
		},
		"relative escape": func() error {
			_, err := Analyze(root, []string{"../outside.go"}, DefaultLimits())
			return err
		},
		"absolute escape": func() error {
			_, err := Analyze(root, []string{filepath.Join(root, "..", "outside.go")}, DefaultLimits())
			return err
		},
		"empty path": func() error {
			_, err := Analyze(root, []string{""}, DefaultLimits())
			return err
		},
		"empty source path": func() error {
			_, err := AnalyzeSource("", []byte("package p\n"), DefaultLimits())
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := run(); err == nil {
				t.Fatal("malformed input was accepted")
			}
		})
	}
}
func TestAnalyzeSourceDoesNotMutateSource(t *testing.T) {
	limits := Limits{MaxFileLines: 300, MaxFunctionLines: 2}
	source := []byte("package p\n\nfunc TooLong() {\n\t_ = 1\n}\n")
	want := append([]byte(nil), source...)
	findings, err := AnalyzeSource("fixture.go", source, limits)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(findings, RuleFunctionLines) {
		t.Fatalf("expected function finding, got %#v", findings)
	}
	if !bytes.Equal(source, want) {
		t.Fatal("AnalyzeSource mutated its source input")
	}
}
