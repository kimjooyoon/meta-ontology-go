package sourcesplit

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestNativeMetricsPreserveUnconnectedGoooCandidate(t *testing.T) {
	var source strings.Builder
	source.WriteString("package fixture\nnamespace fixture\n")
	for index := range 85 {
		fmt.Fprintf(&source, "entity Item%d\n", index)
	}
	root, head := observedRepository(t, map[string]string{
		"large.go":      extractionFixture("Ready"),
		"expression.go": "package fixture\nfunc Value() int {\n value := 1\n return value\n}\n",
		"main.gooo":     source.String(),
	})
	report, raw := observedMetrics(t, root, head)
	observe, err := NewObserver(root, head, raw)
	if err != nil {
		t.Fatal(err)
	}
	plan := generation.BuildWithPrerequisites(head, head, report.Meta, nil, observe)
	if plan.Decision != generation.DecisionPlan || len(plan.Selected) != 2 ||
		plan.Selected[1].Operation != sourcepolicy.OperationExtractFunction ||
		plan.IndicatorDecisionLedger.IndicatorCount != len(report.Meta.Indicators) {
		t.Fatal("native report selected an unconnected operation or erased debt", plan)
	}
	found := false
	for _, prerequisite := range plan.Prerequisites {
		if prerequisite.Reason == "EXECUTION_CONSUMER_UNAVAILABLE" {
			found = true
			if prerequisite.State != "UNKNOWN" || prerequisite.SourceDigest != "" ||
				prerequisite.MetricsDigest != sourceDigest(raw) {
				t.Fatal("consumer observation pretended to inspect source", prerequisite)
			}
		}
	}
	if !found {
		t.Fatal("unconnected Gooo operation has no prerequisite")
	}
	if manifest := generation.BuildExecutionManifest(plan); manifest.Decision != generation.ExecutionDecisionProposed {
		t.Fatal("observed native plan was rejected", manifest)
	}
	if status := observedGit(t, root, "status", "--porcelain", "--untracked-files=all"); status != "" {
		t.Fatal("planning changed the source", status)
	}
}
