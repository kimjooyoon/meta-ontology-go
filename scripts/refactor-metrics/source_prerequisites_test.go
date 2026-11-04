package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestMetricPlanUsesCurrentSourceAndLaterSplittableCandidate(t *testing.T) {
	root, head, source := prerequisiteRepository(t)
	policy := sourcepolicy.Default()
	policy.MaxFileLines, policy.MaxDirectDirectoryIn = 6, 3
	report, err := sourcepolicy.Evaluate(policy, []sourcepolicy.Observation{
		{Subject: "a.go", Dimension: sourcepolicy.DimensionGoFileLines, Value: 8, Producer: "cli-fixture"},
		{Subject: "z.go", Dimension: sourcepolicy.DimensionGoFileLines, Value: 9, Producer: "cli-fixture"},
		{Subject: "z.go:3:First", Dimension: sourcepolicy.DimensionRefactorAssign, Value: 2, Producer: "cli-fixture"},
	})
	if err != nil {
		t.Fatal(err)
	}
	metrics := linecaps.LineMetricsReport{CommitSHA: head, Meta: report,
		Directories: []linecaps.DirectoryMetric{{Path: ".", DirectFiles: 2}}}
	raw, err := json.Marshal(metrics)
	if err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	opts := options{root: root, sha: head, baseSHA: strings.Repeat("a", 40),
		metrics: filepath.Join(output, "metrics.json"), plan: filepath.Join(output, "plan.json")}
	if err := os.WriteFile(opts.metrics, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := emitGenerationPlan(opts, report); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(opts.plan)
	if err != nil {
		t.Fatal(err)
	}
	var plan generation.Plan
	if err := json.Unmarshal(raw, &plan); err != nil {
		t.Fatal(err)
	}
	if plan.Decision != generation.DecisionPlan || len(plan.GoSplitPreparations) != 2 ||
		plan.IndicatorDecisionLedger == nil || plan.IndicatorDecisionLedger.DeferredCount != 1 {
		t.Fatal("CLI omitted native prerequisites or their original debt")
	}
	for _, action := range plan.Selected {
		if action.Operation == sourcepolicy.OperationSplitGo && action.Subject != "z.go" {
			t.Fatal("CLI selected the blocked first file")
		}
	}
	current, err := os.ReadFile(filepath.Join(root, "z.go"))
	if err != nil || string(current) != source {
		t.Fatal("planning rewrote the current source", err)
	}
	opts.sha, opts.plan = strings.Repeat("c", 40), filepath.Join(output, "stale.json")
	if err := emitGenerationPlan(opts, report); err == nil || !strings.Contains(err.Error(), "repository SHA") {
		t.Fatal("stale repository identity acquired a plan", err)
	}
	if _, err := os.Stat(opts.plan); !os.IsNotExist(err) {
		t.Fatal("refusal published a stale plan", err)
	}
}

func prerequisiteRepository(t *testing.T) (string, string, string) {
	t.Helper()
	for _, name := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR"} {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}
	root := t.TempDir()
	const source = "package fixture\n\nfunc First() int {\nreturn 1\n}\n\nfunc Second() int {\nreturn 2\n}\n"
	for name, content := range map[string]string{
		"a.go": "package fixture\nfunc One() int {\nv:=1\nv++\nv++\nv++\nreturn v\n}\n", "z.go": source,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."},
		{"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "core.hooksPath=/dev/null",
			"commit", "--no-gpg-sign", "-qm", "fixture"}} {
		command := exec.Command("git", args...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("prepare local repository: %s: %v", output, err)
		}
	}
	command := exec.Command("git", "rev-parse", "HEAD")
	command.Dir = root
	head, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	return root, strings.TrimSpace(string(head)), source
}
