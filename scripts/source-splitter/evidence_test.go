package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
)

func TestProductionEvidenceConforms(t *testing.T) {
	root := t.TempDir()
	header := "//go:build linux || windows\n\n"
	source := []byte(header + "package sample\n\nimport \"fmt\"\n\n" +
		"var First = fmt.Sprint(\"a\")\nvar Second = \"b\"\n")
	first := []byte(header + "package sample\n\nimport \"fmt\"\n\n" +
		"var First = fmt.Sprint(\"a\")\n")
	second := []byte(header + "package sample\n\nvar Second = \"b\"\n")
	subject := "sample.go"
	target := filepath.Join(root, subject)
	if err := os.WriteFile(target, source, 0o644); err != nil {
		t.Fatal(err)
	}
	plan := splitPlan{Directory: ".", Mode: 0o644, Parts: []splitPart{
		{Path: target, Subject: subject, Data: first},
		{Path: filepath.Join(root, "sample_split02.go"), Subject: "sample_split02.go", Data: second},
	}}
	var output bytes.Buffer
	cfg := config{root: root, sha: strings.Repeat("a", 40), subject: subject, evidenceJSON: true}
	if err := applySplitWithEvidence(cfg, plan, &output); err != nil {
		t.Fatal(err)
	}
	var evidence operationconformance.SplitGoEvidence
	if err := json.Unmarshal(output.Bytes(), &evidence); err != nil {
		t.Fatal(err)
	}
	contract, err := os.ReadFile(filepath.Join("..", "..", "examples",
		"source-splitter-conformance", "contract.json"))
	if err != nil {
		t.Fatal(err)
	}
	report := operationconformance.Evaluate(contract, evidence)
	if report.Decision != operationconformance.DecisionPass ||
		report.Summary.PassCount != 6 || report.Summary.Total != 6 {
		t.Fatalf("decision=%s pass=%d/%d reason=%s", report.Decision,
			report.Summary.PassCount, report.Summary.Total, report.Reason)
	}
}
