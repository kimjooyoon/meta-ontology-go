package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

func TestDecodeExtractorReportAcceptsGeneratedMapConstructionProof(t *testing.T) {
	root, source := mapConstructionConsumerSource(t)
	result, err := projectionextractor.ExtractWithResult(root, "a.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Evidence) != 1 || result.Evidence[0].Strategy != "caller-evaluated-map-construction" {
		t.Fatalf("expected the actual map construction strategy, got %+v", result.Evidence)
	}
	// The envelope uses the existing test-only head. Evidence comes from the
	// compiler, not from renaming a return-tail receipt or inventing digests.
	report := extractionReportWithStrategyEvidence(result.Evidence)
	subject := &report.Subjects[0]
	subject.Before = physicalLineCount([]byte(source))
	subject.After = physicalLineCount(result.Generated["a.go"])
	subject.Files = nil
	for path := range result.Generated {
		subject.Files = append(subject.Files, path)
		if path != "a.go" {
			subject.CreatedFiles = append(subject.CreatedFiles, path)
		}
	}
	sort.Strings(subject.Files)
	sort.Strings(subject.CreatedFiles)
	report.Indicators = extractionTestIndicatorsWithValues(1, 1, 1, len(subject.CreatedFiles), 0)
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "report.json")
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	_, decoded, err := decodeExtractorReport(path, "head")
	if err != nil {
		t.Fatalf("meta executor rejected the generated map proof: %v", err)
	}
	if len(decoded.Subjects) != 1 || !reflect.DeepEqual(decoded.Subjects[0].Evidence, result.Evidence) {
		t.Fatal("meta executor lost generated proof fields")
	}
	if !strings.Contains(decoded.Subjects[0].Evidence[0].ProofStages[3].Detail, "CALLEE_EFFECTS_UNPROVEN") {
		t.Fatal("meta executor lost the previous strategy's unresolved cause")
	}
}

func mapConstructionConsumerSource(t *testing.T) (string, string) {
	t.Helper()
	var source strings.Builder
	source.WriteString("package p\n\nimport \"fmt\"\n\nfunc Observe(next func(string) any) map[string]any {\n")
	source.WriteString(strings.Repeat("\t_ = 1\n", 60))
	source.WriteString("\twitness := map[string]any{\n")
	for index := 1; index <= 16; index++ {
		key := fmt.Sprintf("%02d", index)
		fmt.Fprintf(&source, "\t\t%q: next(%q),\n", key, key)
	}
	source.WriteString("\t}\n\tfmt.Print(len(witness))\n\treturn witness\n}\n")
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module consumer-witness.test\n\ngo 1.27.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(source.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, source.String()
}
