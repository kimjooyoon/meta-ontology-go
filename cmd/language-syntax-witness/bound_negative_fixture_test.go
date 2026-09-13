package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

// Capture the actual read closure, rather than guessing which catalog files matter.
type nativeSyntaxInputs struct {
	fs.FS
	opened map[string]bool
}

func (inputs nativeSyntaxInputs) Open(name string) (fs.File, error) {
	file, err := inputs.FS.Open(name)
	if err == nil {
		inputs.opened[name] = true
	}
	return file, err
}

func nativeBoundSyntaxFixture(t *testing.T) (config, languagesyntax.Report) {
	t.Helper()
	inputs := nativeSyntaxInputs{FS: os.DirFS("../.."), opened: map[string]bool{}}
	registry, err := fs.ReadFile(inputs, "examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	artifact := languageconcept.BuildArtifact(inputs)
	head := strings.Repeat("0", 40)
	baseline := languagesyntax.Evaluate(inputs, head, registry, artifact)
	if err := languagesyntax.Validate(baseline, head); err != nil {
		t.Fatal(err)
	}
	if baseline.Decision != languagesyntax.DecisionPass || !baseline.Source.ConceptBound ||
		!baseline.Source.ObservationKnown || baseline.Summary.Satisfied != languagesyntax.FixedTotal {
		t.Fatalf("original corpus is not a bound positive control: %+v", baseline)
	}
	work := t.TempDir()
	cfg := config{root: filepath.Join(work, "project"), head: head,
		registry: filepath.Join(work, "registry.json"),
		concept:  filepath.Join(work, "concept.json"),
		output:   filepath.Join(work, "report.json")}
	nativeCopySyntaxInputs(t, inputs, cfg.root)
	if err := os.WriteFile(cfg.registry, registry, 0600); err != nil {
		t.Fatal(err)
	}
	nativeWriteSyntaxConcept(t, cfg, artifact)
	return cfg, baseline
}

func nativeCopySyntaxInputs(t *testing.T, inputs nativeSyntaxInputs, root string) {
	t.Helper()
	names := make([]string, 0, len(inputs.opened))
	for name := range inputs.opened {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		info, err := fs.Stat(inputs.FS, name)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(root, filepath.FromSlash(name))
		if info.IsDir() {
			if err := os.MkdirAll(target, 0700); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if !info.Mode().IsRegular() {
			t.Fatal(fmt.Errorf("unsupported syntax input: %s", name))
		}
		raw, err := fs.ReadFile(inputs.FS, name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func nativeWriteSyntaxConcept(t *testing.T, cfg config, artifact languageconcept.Artifact) {
	t.Helper()
	raw, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg.concept, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

func nativeReadBoundSyntaxReport(t *testing.T, cfg config) (languagesyntax.Report, []byte) {
	t.Helper()
	raw, err := os.ReadFile(cfg.output)
	if err != nil {
		t.Fatal(err)
	}
	var report languagesyntax.Report
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if err := languagesyntax.Validate(report, cfg.head); err != nil {
		t.Fatal(err)
	}
	return report, raw
}
