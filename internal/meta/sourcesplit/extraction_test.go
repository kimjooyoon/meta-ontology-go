package sourcesplit

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	recipeauthority "github.com/kimjooyoon/meta-ontology-go/internal/meta/functionextractorrecipe"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func extractionFixture(name string) string {
	return "package fixture\nfunc " + name + "(values map[string]struct{}) error {\n" +
		strings.Repeat(" _ = 1\n", 72) + " if len(values) != 0 { return nil }\n return nil\n}\n"
}

func TestExtractionPrerequisitePreservesNativeFailureAndSelectsLaterCandidate(t *testing.T) {
	root, head := observedRepository(t, map[string]string{
		"go.mod": "module example.test/fixture\n",
		"a.go": "package fixture\nimport \"io\"\nfunc Blocked(w io.Writer) error {\n" +
			strings.Repeat(" _ = 1\n", 72) + " _, err := io.WriteString(w, \"x\")\n return err\n}\n",
		"b.go":          extractionFixture("Ready"),
		"expression.go": "package fixture\nfunc Value() int {\n value := 1\n return value\n}\n",
	})
	report, raw := observedMetrics(t, root, head)
	observe, err := NewObserver(root, head, raw)
	if err != nil {
		t.Fatal(err)
	}
	plan := generation.BuildWithPrerequisites(head, head, report.Meta, nil, observe)
	if plan.Decision != generation.DecisionPlan || len(plan.Selected) != 2 ||
		plan.Selected[1].Operation != sourcepolicy.OperationExtractFunction ||
		!strings.HasPrefix(plan.Selected[1].Subject, "b.go:") {
		t.Fatal("native extraction failure hid the later candidate", plan)
	}
	found := false
	for _, prerequisite := range plan.Prerequisites {
		if prerequisite.Reason == "CALLEE_EFFECTS_UNPROVEN" {
			found = true
			if prerequisite.State != "UNKNOWN" || prerequisite.UnknownClass != "DIRECT_MISSING" ||
				prerequisite.Step != "prove-callee-effects" || prerequisite.NextOperation != "restore-callee-evidence" ||
				len(prerequisite.Diagnostics) == 0 || prerequisite.SourceDigest == "" {
				t.Fatal("native failure lost its classification or source", prerequisite)
			}
		}
	}
	if !found || plan.IndicatorDecisionLedger.IndicatorCount != len(report.Meta.Indicators) ||
		!reflect.DeepEqual(plan, generation.BuildWithPrerequisites(head, head, report.Meta, nil, observe)) {
		t.Fatal("native failure or source denominator disappeared, or replay changed", plan)
	}
	if manifest := generation.BuildExecutionManifest(plan); manifest.Decision != generation.ExecutionDecisionProposed {
		t.Fatal("observed extraction plan is not canonical", manifest)
	}
	if status := observedGit(t, root, "status", "--porcelain", "--untracked-files=all"); status != "" {
		t.Fatal("read-only preparation changed source", status)
	}
}

func TestExtractionPrerequisiteUsesActiveRecipeBeforeGenericPlanning(t *testing.T) {
	source := "package fixture\nfunc Value() int { return 1 }\n"
	root, head := observedRepository(t, map[string]string{"source.go": source})
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		t.Fatal(err)
	}
	action := generation.Action{Operation: sourcepolicy.OperationExtractFunction, Subject: "source.go:2:Value",
		IndicatorID: "sha256:" + strings.Repeat("a", 64), InputContractSourceDigest: contract.SourceDigest,
		InputContractSemanticDigest: contract.SemanticDigest}
	recipe := recipeauthority.ExtractionRecipe{Subject: "source.go", Operation: "extract-function",
		Edits:   []recipeauthority.TextEdit{{Path: "source.go", Old: []string{"return 1"}, New: []string{"return Helper()"}}},
		Creates: []recipeauthority.FileCreation{{Path: "helper.go", Lines: []string{"package fixture", "func Helper() int { return 1 }"}}}}
	observed := observeExtraction(root, head, strings.Repeat("d", 64), action, []recipeauthority.ExtractionRecipe{recipe}, nil)
	if observed.State != "READY" || observed.Reason != "EXTRACTION_PLAN_READY" {
		t.Fatal("active recipe was replaced by an unrelated generic result", observed)
	}
	if raw, err := os.ReadFile(filepath.Join(root, "source.go")); err != nil || string(raw) != source {
		t.Fatal("preparation rewrote source", err)
	}
	if _, err := os.Stat(filepath.Join(root, "helper.go")); !os.IsNotExist(err) {
		t.Fatal("preparation installed a helper", err)
	}
}
