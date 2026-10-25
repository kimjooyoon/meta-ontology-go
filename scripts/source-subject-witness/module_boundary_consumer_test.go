package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func boundaryModuleReport(t *testing.T, mode string) sourceReport {
	t.Helper()
	return nativeModuleReportWithSetup(t, func(root string) {
		manifest := filepath.Join(root, "nested", "go.mod")
		var err error
		switch mode {
		case "toolchain":
			err = os.WriteFile(manifest, []byte("module example.test/nested\ngo 1.99.0\n"), 0644)
		case "owner":
			if err = os.Remove(manifest); err == nil {
				err = os.Symlink("../go.mod", manifest)
			}
		case "cross-module":
			source := filepath.Join(root, "nested", "source.go")
			if err = os.WriteFile(filepath.Join(root, "root.go"), []byte("package root\nfunc Increment(x int) int { return x + 1 }\n"+strings.Repeat("// line\n", 75)), 0644); err == nil {
				if err = os.Remove(source); err == nil {
					err = os.Symlink("../root.go", source)
				}
			}
		}
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestNativeModuleBoundaryKeepsSourceObservation(t *testing.T) {
	for mode, reason := range map[string]string{"toolchain": "GO_MODULE_TOOLCHAIN_UNAVAILABLE", "owner": "GO_MODULE_OWNER_UNAVAILABLE", "cross-module": "GO_SOURCE_CROSSES_MODULE_BOUNDARY"} {
		t.Run(mode, func(t *testing.T) {
			report := boundaryModuleReport(t, mode)
			ledger, err := buildLedger(report, report.CommitSHA)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, w := range ledger.Witnesses {
				if w.Path == "nested/source.go" {
					found = true
					if w.ModuleBoundaryReason != reason || w.Metrics[0].Value != 77 || w.Meta.NotApplicableIndicators != 1 {
						t.Fatal(w)
					}
					changed := w
					changed.ModuleBoundaryReason = ""
					if sealWitness(changed).WitnessDigest == w.WitnessDigest {
						t.Fatal("reason omitted from identity")
					}
				}
			}
			if !found || ledger.Counts.GoFiles != 2 {
				t.Fatal("source was dropped", ledger.Counts)
			}
			replay, err := buildLedger(report, report.CommitSHA)
			if err != nil || digestJSON(ledger) != digestJSON(replay) {
				t.Fatal("replay", err)
			}
		})
	}
}

func TestModuleBoundaryContradictionsReturnNoLedger(t *testing.T) {
	for _, mutation := range []string{"missing-file-reason", "unknown-file-reason", "indicator-reason", "forged-pass", "failure-code"} {
		t.Run(mutation, func(t *testing.T) {
			report := boundaryModuleReport(t, "toolchain")
			for i := range report.Files {
				f := &report.Files[i]
				if f.Path == "nested/source.go" {
					if mutation == "missing-file-reason" {
						f.ModuleBoundaryReason = ""
					}
					if mutation == "unknown-file-reason" {
						f.ModuleBoundaryReason = "UNKNOWN"
					}
				}
			}
			for i := range report.Meta.Indicators {
				r := &report.Meta.Indicators[i]
				if r.Subject == "nested/source.go" {
					switch mutation {
					case "indicator-reason":
						r.ApplicabilityReason = "GO_MODULE_OWNER_UNAVAILABLE"
					case "forged-pass":
						r.Decision = "PASS"
					case "failure-code":
						r.FailureCode = ""
					}
				}
			}
			ledger, err := buildLedger(report, report.CommitSHA)
			if err == nil || ledger.Status != "" {
				t.Fatal("contradiction accepted", err)
			}
		})
	}
}
