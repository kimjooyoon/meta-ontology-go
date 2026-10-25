package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metabinding"
)

func TestNativeNestedModuleReportRetainsConsumerOrigins(t *testing.T) {
	report := nativeModuleReport(t)
	ledger, err := buildLedger(report, report.CommitSHA)
	if err != nil {
		t.Fatal(err)
	}
	nested := 0
	for _, witness := range ledger.Witnesses {
		if witness.ModuleRoot == "nested" {
			nested++
			if witness.Path != "nested/source.go" && witness.Path != "nested/source.go:2:Increment" {
				t.Fatalf("unexpected owned subject: %+v", witness)
			}
			if witness.Space == "LOGICAL_FILE" && witness.Metrics[0].Value != 77 {
				t.Fatal("source line count changed", witness.Metrics)
			}
		}
	}
	if nested != 2 || ledger.Counts.GoFiles != 2 || ledger.ClaimState != "OPEN" || len(ledger.SourceObservations) != 1 {
		t.Fatalf("module consumer coverage changed: nested=%d counts=%+v observations=%+v", nested, ledger.Counts, ledger.SourceObservations)
	}
	if ledger.SourceObservations[0].ModuleRoot != "nested" || ledger.SourceObservations[0].Satisfied {
		t.Fatal("unsatisfied nested observation lost its origin", ledger.SourceObservations)
	}
	replay, err := buildLedger(report, report.CommitSHA)
	if err != nil || digestJSON(ledger) != digestJSON(replay) {
		t.Fatal("replay changed", err)
	}
	for _, original := range ledger.Witnesses {
		if original.ModuleRoot != "nested" {
			continue
		}
		changed := original
		changed.ModuleRoot = ""
		if sealWitness(changed).WitnessDigest == original.WitnessDigest {
			t.Fatal("owner omitted from digest")
		}
	}
}

func TestModuleOriginDisagreementsRejectNativeReport(t *testing.T) {
	for _, mutation := range []string{"file-owner", "indicator-owner", "function-owner", "outside-module"} {
		t.Run(mutation, func(t *testing.T) {
			report := nativeModuleReport(t)
			for i := range report.Files {
				file := &report.Files[i]
				if file.Path == "nested/source.go" {
					if mutation == "file-owner" {
						file.ModuleRoot = ""
					}
					if mutation == "outside-module" {
						file.ModuleRoot = "other"
					}
				}
			}
			for i := range report.Meta.Indicators {
				row := &report.Meta.Indicators[i]
				if mutation == "indicator-owner" && row.Subject == "nested/source.go" {
					row.ModuleRoot = ""
				}
				if mutation == "function-owner" && row.SubjectKind == "FUNCTION" {
					row.ModuleRoot = ""
				}
			}
			if ledger, err := buildLedger(report, report.CommitSHA); err == nil || ledger.Status != "" {
				t.Fatal("inconsistent module origin produced a ledger", err)
			}
		})
	}
}

func nativeModuleReport(t *testing.T) sourceReport {
	t.Helper()
	root, storage := t.TempDir(), t.TempDir()
	inputs := map[string]string{
		".github/workflows/ci.yml": "name: fixture\n",
		"go.mod":                   "module example.test/root\ngo 1.27.3\n",
		"root.go":                  "package root\n",
		"nested/go.mod":            "module example.test/nested\ngo 1.27.3\n",
		"nested/source.go":         "package nested\nfunc Increment(x int) int { return x + 1 }\n" + strings.Repeat("// line\n", 75),
	}
	for path, content := range inputs {
		target := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	workflow := filepath.Join(storage, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(filepath.Dir(workflow), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(workflow, []byte("name: fixture\n"), 0644); err != nil {
		t.Fatal(err)
	}
	produced, err := linecaps.AnalyzeProjectedLineMetrics(root, storage)
	if err != nil {
		t.Fatal(err)
	}
	produced.Repository, produced.CommitSHA = "example.test/origins", strings.Repeat("a", 40)
	raw, err := json.Marshal(produced)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "metrics.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	input, err := metabinding.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	_, bound, err := metabinding.Build(filepath.Join("..", ".."), input)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = metabinding.Augment(input, bound)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	report, err := loadSource(path)
	if err != nil {
		t.Fatal(err)
	}
	return report
}
