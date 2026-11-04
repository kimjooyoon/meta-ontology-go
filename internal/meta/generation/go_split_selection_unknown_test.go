package generation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestSourceSelectionKeepsMissingDirectoryUnknownAndUsesLaterFile(t *testing.T) {
	root, report, topology := sourceSelectionFixture(t, false)
	for _, directory := range []string{"a", "z"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, directory, "fixture.go"), []byte(splitPlannerSource), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	for index := range report.Indicators {
		switch report.Indicators[index].Subject {
		case "a.go":
			report.Indicators[index].Subject = "a/fixture.go"
		case "z.go":
			report.Indicators[index].Subject = "z/fixture.go"
		}
	}
	topology.Directories = []metricevidence.Directory{{Path: "z", DirectFiles: 1}}
	plan, err := BuildWithSourcePrerequisites(root, strings.Repeat("a", 40), topology.CommitSHA, report, topology, OperationInputs{})
	if err != nil || plan.Decision != DecisionPlan || !receiptPlanKnown(plan) || len(plan.GoSplitPreparations) != 2 {
		t.Fatal("missing directory stopped independent selection", err)
	}
	found := false
	for _, observation := range plan.GoSplitPreparations {
		if observation.States[1] == PrerequisiteUnknown {
			found = true
			if observation.NextStep().F_next_operation != "observe-directory" {
				t.Fatal("unknown directory acquired a repair conclusion")
			}
		}
	}
	for _, action := range plan.Selected {
		if action.Operation == sourcepolicy.OperationSplitGo && action.Subject != "z/fixture.go" {
			t.Fatal("unobserved directory acquired a selected rewrite")
		}
	}
	if !found || plan.IndicatorDecisionLedger.DeferredCount != 1 {
		t.Fatal("missing observation or original metric debt disappeared")
	}
}

func TestSourceSelectionRetainsInspectionWhenAnotherOperationInputIsMissing(t *testing.T) {
	root, report, topology := sourceSelectionFixture(t, false)
	registration, _ := registrationPlanFixture(t)
	for index := range report.Indicators {
		if report.Indicators[index].Operation == sourcepolicy.OperationCollapseAssign {
			report.Indicators[index] = registration.Indicators[0]
		}
	}
	plan, err := BuildWithSourcePrerequisites(root, strings.Repeat("a", 40), topology.CommitSHA, report, topology, OperationInputs{})
	if err != nil || plan.Decision != DecisionUnknown || plan.Reason != ReasonRegistrationInput ||
		len(plan.Selected) != 0 || len(plan.GoSplitPreparations) != 2 || !receiptPlanKnown(plan) ||
		validatePlanIndicatorDecisionLedger(plan) != nil {
		t.Fatalf("unbound input erased inspections or invalidated their ledger: %+v, %v", plan, err)
	}
	if plan.IndicatorDecisionLedger.IndicatorCount != 3 || plan.IndicatorDecisionLedger.DeferredCount != 3 ||
		len(plan.OperationInputFailures)+len(plan.RegistrationInputFailures) != 1 {
		t.Fatal("unbound input lost its separate failure or original denominator")
	}
}
