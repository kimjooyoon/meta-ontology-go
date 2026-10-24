package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestExtractionFailureReturnsBeforeMaterialization(t *testing.T) {
	root := t.TempDir()
	source := "package fixture\nimport \"io\"\nfunc Blocked(w io.Writer) error {\n" +
		strings.Repeat(" _ = 1\n", 72) + " _, err := io.WriteString(w, \"x\")\n return err\n}\n"
	for name, data := range map[string]string{"go.mod": "module example.test/fixture\n", "fixture.go": source} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"commit", "-qm", "fixture"}} {
		registrationTestCommand(t, root, "git", append([]string{"-c", "user.name=Fixture",
			"-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
	}
	head := collapseTestHead(t, root)
	report := metricevidence.Report{CommitSHA: head}
	report.Meta.Policy.MaxFileLines, report.Meta.Policy.MaxDirectEntries = 75, 10
	metrics := filepath.Join(t.TempDir(), "metrics.json")
	registrationTestWriteJSON(t, metrics, report)
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		t.Fatal(err)
	}
	action := generation.Action{Operation: sourcepolicy.OperationExtractFunction, Subject: "fixture.go:3:Blocked",
		IndicatorID: "sha256:" + strings.Repeat("a", 64), InputContractSourceDigest: contract.SourceDigest,
		InputContractSemanticDigest: contract.SemanticDigest}
	materialized, failure := executeAction(root, "", metrics, generation.Plan{HeadSHA: head}, action, metaExecutionTrace{})
	if failure == nil || failure.reason != "CALLEE_EFFECTS_UNPROVEN" || failure.step != "prove-callee-effects" ||
		failure.class != "DIRECT_MISSING" || failure.next != "restore-callee-evidence" ||
		!slices.Contains(failure.diagnostics, "callee=io.WriteString") || len(materialized.Executor.Command) != 0 {
		t.Fatal("native failure was lost or reached the writer", failure, materialized)
	}
	if status := registrationTestCommand(t, root, "git", "status", "--porcelain", "--untracked-files=all"); len(status) != 0 {
		t.Fatal("refused extraction changed source", string(status))
	}
}
