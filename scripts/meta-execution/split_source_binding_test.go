package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestSplitExecutionBindsPreparedBytesBeforeGitAndProcesses(t *testing.T) {
	root := t.TempDir()
	const source = "package fixture\n\nfunc First() int {\nreturn 1\n}\n\nfunc Second() int {\nreturn 2\n}\n"
	file := filepath.Join(root, "fixture.go")
	if err := os.WriteFile(file, []byte(source), 0o640); err != nil {
		t.Fatal(err)
	}
	policy := sourcepolicy.Default()
	policy.MaxFileLines, policy.MaxDirectDirectoryIn = 6, 2
	report, err := sourcepolicy.Evaluate(policy, []sourcepolicy.Observation{
		{Subject: "fixture.go", Dimension: sourcepolicy.DimensionGoFileLines, Value: 9, Producer: "execution-fixture"},
		{Subject: "fixture.go:3:First", Dimension: sourcepolicy.DimensionRefactorAssign, Value: 2, Producer: "execution-fixture"},
	})
	if err != nil {
		t.Fatal(err)
	}
	topology := metricevidence.Report{CommitSHA: strings.Repeat("b", 40),
		Directories: []metricevidence.Directory{{Path: ".", DirectFiles: 1}}}
	topology.Meta.Policy.MaxFileLines, topology.Meta.Policy.MaxDirectEntries = 6, 2
	plan, err := generation.BuildWithSourcePrerequisites(root, strings.Repeat("a", 40), topology.CommitSHA,
		report, topology, generation.OperationInputs{})
	if err != nil || plan.Decision != generation.DecisionPlan {
		t.Fatal("native plan unavailable", err)
	}
	for _, action := range plan.Selected {
		if action.Operation != sourcepolicy.OperationSplitGo {
			continue
		}
		_, current := materializeSplit(root, "unavailable-git-directory", "unused-metrics", plan, action,
			metaExecutionTrace{}, "first")
		if current == nil || current.reason != "GIT_SNAPSHOT_UNAVAILABLE" {
			t.Fatalf("current source failed before the deliberately absent Git input: %+v", current)
		}
		changed := "//새 소스\n" + source
		if err := os.WriteFile(file, []byte(changed), 0o640); err != nil {
			t.Fatal(err)
		}
		materialized, stale := materializeSplit(root, "unavailable-git-directory", "unused-metrics", plan, action,
			metaExecutionTrace{}, "first")
		if stale == nil || stale.reason != "GO_SPLIT_PREREQUISITE_STALE" || stale.stage != "prepare-workspace" ||
			len(materialized.Executor.Command) != 0 {
			t.Fatalf("changed source reached Git or an executor: materialized=%+v failure=%+v", materialized, stale)
		}
		raw, err := os.ReadFile(file)
		if err != nil || string(raw) != changed {
			t.Fatal("refusal changed the authored source", err)
		}
		entries, err := os.ReadDir(root)
		if err != nil || len(entries) != 1 {
			t.Fatal("refusal published files or temporaries", err)
		}
		return
	}
	t.Fatal("no split action selected")
}
