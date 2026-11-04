package generation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestSourcePrerequisiteSelectionUsesLaterCandidateWithoutErasingDebt(t *testing.T) {
	root, report, topology := sourceSelectionFixture(t, false)
	before, _ := json.Marshal(report)
	first, err := BuildWithSourcePrerequisites(root, strings.Repeat("a", 40), topology.CommitSHA, report, topology, OperationInputs{})
	if err != nil || first.Decision != DecisionPlan || len(first.Selected) != 2 || len(first.GoSplitPreparations) != 2 {
		t.Fatalf("later native candidate was not selected: %+v, %v", first, err)
	}
	second, err := BuildWithSourcePrerequisites(root, strings.Repeat("a", 40), topology.CommitSHA, report, topology, OperationInputs{})
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatal("same source-bound plan did not replay", err)
	}
	selected := ""
	for _, action := range first.Selected {
		if action.Operation == sourcepolicy.OperationSplitGo {
			selected = action.Subject
		}
	}
	if selected != "z.go" || first.IndicatorDecisionLedger == nil || first.IndicatorDecisionLedger.IndicatorCount != 3 ||
		first.IndicatorDecisionLedger.DeferredCount != 1 || !receiptPlanKnown(first) || validatePlanIndicatorDecisionLedger(first) != nil {
		t.Fatalf("native refusal changed group ownership or denominator: %+v", first)
	}
	after, _ := json.Marshal(report)
	if string(before) != string(after) || first.IndicatorsDigest != digestJSON(normalizeIndicators(report.Indicators)) {
		t.Fatal("selection rewrote original metric pressure")
	}
}

func TestSourcePrerequisiteShortfallRetainsAllUnmetIndicators(t *testing.T) {
	root, report, topology := sourceSelectionFixture(t, true)
	plan, err := BuildWithSourcePrerequisites(root, strings.Repeat("a", 40), topology.CommitSHA, report, topology, OperationInputs{})
	if err != nil || plan.Decision != DecisionUnknown || plan.Reason != ReasonPressureShortfall || len(plan.Selected) != 0 ||
		plan.IndicatorDecisionLedger == nil || plan.IndicatorDecisionLedger.IndicatorCount != 3 ||
		plan.IndicatorDecisionLedger.DeferredCount != 3 || !receiptPlanKnown(plan) {
		t.Fatalf("unmet split candidates became a fixed point or disappeared: %+v, %v", plan, err)
	}
	for _, entry := range plan.IndicatorDecisionLedger.Entries {
		if entry.SourceIndicator.Satisfied {
			t.Fatal("unmet original observation was changed to satisfied")
		}
	}
}

func TestSourcePrerequisiteManifestRefusesFalseCompletionAndMissingSource(t *testing.T) {
	root, report, topology := sourceSelectionFixture(t, false)
	plan, err := BuildWithSourcePrerequisites(root, strings.Repeat("a", 40), topology.CommitSHA, report, topology, OperationInputs{})
	if err != nil || BuildExecutionManifest(plan).Decision != ExecutionDecisionProposed {
		t.Fatal("positive plan did not enter the existing manifest", err)
	}
	for name, mutate := range map[string]func(*Plan){
		"policy":            func(p *Plan) { p.GoSplitPolicySHA256 = strings.Repeat("0", 64) },
		"false conformance": func(p *Plan) { p.GoSplitPreparations[0].States[2] = PrerequisiteObserved },
		"missing raw source": func(p *Plan) {
			for i := range p.GoSplitPreparations {
				p.GoSplitPreparations[i].SourceSHA256 = ""
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			changed := plan
			changed.GoSplitPreparations = append([]GoSplitObservation{}, plan.GoSplitPreparations...)
			mutate(&changed)
			changed = finish(changed)
			if BuildExecutionManifest(changed).Decision != ExecutionDecisionUnknown {
				t.Fatal("self-resealed false prerequisite entered execution")
			}
		})
	}
}

func TestPreparedGoSplitSourceRejectsChangedBytesAndAnotherOccurrence(t *testing.T) {
	root, report, topology := sourceSelectionFixture(t, false)
	plan, err := BuildWithSourcePrerequisites(root, strings.Repeat("a", 40), topology.CommitSHA, report, topology, OperationInputs{})
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range plan.Selected {
		if action.Operation != sourcepolicy.OperationSplitGo {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, action.Subject))
		if err != nil || ValidatePreparedGoSplitSource(plan, action, raw) != nil {
			t.Fatal("current source rejected", err)
		}
		if ValidatePreparedGoSplitSource(plan, action, append([]byte("//changed\n"), raw...)) == nil {
			t.Fatal("changed current source reused the preparation")
		}
		action.Subject = "same-bytes-elsewhere.go"
		if ValidatePreparedGoSplitSource(plan, action, raw) == nil {
			t.Fatal("another source occurrence borrowed the preparation")
		}
		return
	}
	t.Fatal("no selected split")
}

func sourceSelectionFixture(t *testing.T, bothBlocked bool) (string, sourcepolicy.Report, metricevidence.Report) {
	t.Helper()
	root := t.TempDir()
	blocked := "package fixture\nfunc One() int {\nv:=1\nv++\nv++\nv++\nreturn v\n}\n"
	later := splitPlannerSource
	if bothBlocked {
		later = blocked
	}
	for name, raw := range map[string]string{"a.go": blocked, "z.go": later} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(raw), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	policy := sourcepolicy.Default()
	policy.MaxFileLines, policy.MaxDirectDirectoryIn = 6, 3
	report, err := sourcepolicy.Evaluate(policy, []sourcepolicy.Observation{
		{Subject: "a.go", Dimension: sourcepolicy.DimensionGoFileLines, Value: 8, Producer: "native-fixture"},
		{Subject: "z.go", Dimension: sourcepolicy.DimensionGoFileLines, Value: 11, Producer: "native-fixture"},
		{Subject: "c.go:3:Collapse", Dimension: sourcepolicy.DimensionRefactorAssign, Value: 2, Producer: "native-fixture"},
	})
	if err != nil {
		t.Fatal(err)
	}
	topology := metricevidence.Report{CommitSHA: strings.Repeat("b", 40),
		Directories: []metricevidence.Directory{{Path: ".", DirectFiles: 2}}}
	topology.Meta.Policy.MaxFileLines, topology.Meta.Policy.MaxDirectEntries = policy.MaxFileLines, policy.MaxDirectDirectoryIn
	return root, report, topology
}
