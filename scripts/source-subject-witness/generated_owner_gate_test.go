package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The six companion documents are shape fixtures, not CI acceptance evidence.
func ownerGateFixture(t *testing.T, report sourceReport) string {
	t.Helper()
	root := t.TempDir()
	sha := "sha256:" + strings.Repeat("a", 64)
	bare := strings.Repeat("a", 64)
	documents := map[string]any{
		"source-metrics":              report,
		"meta-binding-report":         map[string]any{"schema": "gooo/meta-binding-report/v1", "decision": "PASS", "summary": map[string]int{"unbound_indicators": 0}, "witnesses": []any{}, "report_digest": sha},
		"meta-actionability-report":   map[string]any{"schema": "gooo/meta-actionability-report/v1", "indicators": []any{}, "operations": []any{}, "report_digest": sha},
		"self-improvement-plan":       map[string]any{"schema_version": "gooo/self-improvement-generation/v7", "decision": "FIXED_POINT", "registry": []any{}, "selected": []any{}, "unknown_indicator_ids": []any{}, "refuted_indicator_ids": []any{}, "counterexamples": []any{}, "promotion_authorized": false, "plan_digest": bare, "replay_digest": bare},
		"self-improvement-execution":  map[string]any{"schema_version": "gooo/meta-operation-execution/v7", "decision": "FIXED_POINT", "steps": []any{}, "promotion_authorized": false, "manifest_digest": bare, "replay_digest": bare},
		"self-improvement-receipts":   map[string]any{"schema_version": "gooo/meta-operation-receipt-report/v3", "decision": "FIXED_POINT", "receipts": []any{}, "failures": []any{}, "unknowns": []any{}, "promotion_authorized": false, "report_digest": bare, "replay_digest": bare},
		"self-improvement-provenance": map[string]any{"schema_version": "gooo/meta-artifact-provenance/v1", "decision": "BOUND", "summary": map[string]int{"fail": 0, "unknown": 0, "pass": 0}, "indicators": []any{}, "promotion_authorized": false, "envelope_digest": bare, "replay_digest": bare},
	}
	for name, document := range documents {
		raw, err := json.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
		replay := name + "-replay"
		if name == "self-improvement-plan" {
			replay = "self-improvement-replay"
		}
		for _, output := range []string{name, replay} {
			if err := os.WriteFile(filepath.Join(root, output+".json"), raw, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	return root
}

func TestCIGateUsesNativeSourceOwnedApplicability(t *testing.T) {
	for _, command := range []string{"bash", "jq"} {
		if _, err := exec.LookPath(command); err != nil {
			t.Skip("CI gate requires", command)
		}
	}
	report := generatedOwnerReport(t)
	for _, corrupt := range []bool{false, true} {
		t.Run(map[bool]string{false: "retained-owner", true: "unbound-owner"}[corrupt], func(t *testing.T) {
			copy := report
			copy.Files = append([]fileMetric(nil), report.Files...)
			if corrupt {
				for i := range copy.Files {
					copy.Files[i].GeneratorOwned = false
				}
			}
			root := ownerGateFixture(t, copy)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			command := exec.CommandContext(ctx, "bash", "scripts/verify/validate-source-observations.sh", root)
			command.Dir = filepath.Join("..", "..")
			command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOOO_CI_SOURCE_HEAD="+report.CommitSHA, "GITHUB_STEP_SUMMARY="+filepath.Join(root, "summary.md"), "PATH="+filepath.Join(runtime.GOROOT(), "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
			raw, err := command.CombinedOutput()
			if !corrupt && err != nil {
				t.Fatal(err, string(raw))
			}
			if corrupt && (err == nil || !strings.Contains(string(raw), "SOURCE_GENERATOR_OWNER_CONTRADICTION")) {
				t.Fatal("gate did not use the source-owned correspondence check", err, string(raw))
			}
		})
	}
}
