package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcesplit"
)

func TestSplitExecutionRechecksPrerequisiteOriginsBeforeWriting(t *testing.T) {
	root := t.TempDir()
	source := []byte("package fixture\n\nfunc First() int {\n return 1\n}\n\nfunc Second() int {\n return 2\n}\n")
	path, metrics := filepath.Join(root, "fixture.go"), filepath.Join(t.TempDir(), "metrics.json")
	if err := os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"commit", "-qm", "fixture"}} {
		registrationTestCommand(t, root, "git", append([]string{"-c", "user.name=Fixture",
			"-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
	}
	head := collapseTestHead(t, root)
	report := metricevidence.Report{CommitSHA: head, Directories: []metricevidence.Directory{{Path: ".", DirectFiles: 1}}}
	report.Meta.Policy.MaxFileLines, report.Meta.Policy.MaxDirectEntries = 5, 10
	contract, err := generation.SplitGoInputContractEvidence()
	if err != nil {
		t.Fatal(err)
	}
	action := generation.Action{Operation: sourcepolicy.OperationSplitGo, Subject: "fixture.go",
		IndicatorID: "sha256:" + strings.Repeat("a", 64), InputContractSourceDigest: contract.SourceDigest,
		InputContractSemanticDigest: contract.SemanticDigest}
	raw, _ := json.Marshal(report)
	if err := os.WriteFile(metrics, raw, 0600); err != nil {
		t.Fatal(err)
	}
	observe, err := sourcesplit.NewObserver(root, head, raw)
	if err != nil {
		t.Fatal(err)
	}
	plan := generation.Plan{HeadSHA: head, Prerequisites: []generation.CandidatePrerequisite{observe(action)}}
	if failure := checkSourcePrerequisite(root, metrics, plan, action); failure != nil {
		t.Fatal("unchanged input was rejected", failure)
	}
	if !sourceMatchesPrerequisite(plan, action, source) || sourceMatchesPrerequisite(plan, action, append(source, '\n')) {
		t.Fatal("execution source bytes were not bound to the preflight digest")
	}
	if err := os.WriteFile(path, append(source, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	_, failure := executeAction(root, "", metrics, plan, action, metaExecutionTrace{})
	if failure == nil || failure.reason != "PREREQUISITE_INPUT_CHANGED" || failure.class != "STALE" {
		t.Fatal("changed source reached materialization", failure)
	}
	if err := os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
	}
	checkLegacySplitBlockers(t, root, metrics, plan, action, report)
	if status := registrationTestCommand(t, root, "git", "status", "--porcelain", "--untracked-files=all"); len(status) != 0 {
		t.Fatal("refusal changed source", string(status))
	}
}

func checkLegacySplitBlockers(t *testing.T, root, metrics string, plan generation.Plan, action generation.Action, report metricevidence.Report) {
	t.Helper()
	plan.Prerequisites = nil // Old plans must still pass the current executor preflight.
	for _, tc := range []struct{ reason, class string }{
		{"LOGICAL_DIRECTORY_CAPACITY", "KNOWN_CONTRADICTION"},
		{"SUCCESSOR_PROJECTION_REQUIRED", "DEPENDENCY_BLOCKED"},
		{"DIRECTORY_OBSERVATION_MISSING", "DIRECT_MISSING"},
	} {
		copy := report
		switch tc.reason {
		case "LOGICAL_DIRECTORY_CAPACITY":
			copy.Directories = []metricevidence.Directory{{Path: ".", DirectFiles: 10}}
		case "SUCCESSOR_PROJECTION_REQUIRED":
			copy.StorageRoot = "physical"
		case "DIRECTORY_OBSERVATION_MISSING":
			copy.Directories = nil
		}
		registrationTestWriteJSON(t, metrics, copy)
		_, failure := executeAction(root, "", metrics, plan, action, metaExecutionTrace{})
		if failure == nil || failure.reason != tc.reason || failure.class != tc.class {
			t.Fatalf("typed prerequisite was lost before execution: %+v", failure)
		}
	}
}
