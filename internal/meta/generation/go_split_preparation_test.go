package generation

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/gosplit"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestGoSplitPreparationKeepsNativeAndSourceScopesDistinct(t *testing.T) {
	for _, test := range []struct {
		name, source string
		operation    string
		directories  []metricevidence.Directory
		kind         gosplit.BlockerKind
		first, next  PrerequisiteState
	}{
		{"candidate", splitPlannerSource, "verify-projected-workspace",
			[]metricevidence.Directory{{Path: ".", DirectFiles: 1}}, "",
			PrerequisiteObserved, PrerequisiteObserved},
		{"rendered unavailable", "package fixture\nfunc One() int {\nv:=1\nv++\nv++\nv++\nreturn v\n}\n",
			"extract-function",
			[]metricevidence.Directory{{Path: ".", DirectFiles: 1}}, gosplit.RenderedPlanUnavailable,
			PrerequisiteUnmet, PrerequisiteUnobserved},
		{"no split", "package fixture\nfunc One() int { return 1 }\n",
			"refresh-source-metrics",
			[]metricevidence.Directory{{Path: ".", DirectFiles: 1}}, gosplit.NoDeclarationSplit,
			PrerequisiteUnmet, PrerequisiteUnobserved},
		{"full directory", splitPlannerSource, "repack-directory", []metricevidence.Directory{{Path: ".", DirectFiles: 2}},
			gosplit.DirectoryFull, PrerequisiteObserved, PrerequisiteUnmet},
		{"mixed directory", splitPlannerSource, "repack-directory",
			[]metricevidence.Directory{{Path: ".", DirectFiles: 1, DirectFolders: 1}},
			gosplit.DirectoryMixed, PrerequisiteObserved, PrerequisiteUnmet},
		{"missing observation", splitPlannerSource, "observe-directory", nil, gosplit.DirectoryUnobserved,
			PrerequisiteObserved, PrerequisiteUnknown},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, report := splitPlannerFixture(t, test.source)
			report.Directories = test.directories
			report.Meta.Indicators = []metricevidence.Indicator{{Subject: "fixture.go",
				MetricID: sourcepolicy.DimensionGoFileLines, Satisfied: false}}
			before, _ := json.Marshal(report)
			result, err := InspectGoSplit(root, "fixture.go", report)
			if (err == nil) != (test.kind == "") {
				t.Fatalf("native outcome changed: %v", err)
			}
			if test.kind != "" {
				var blocker *gosplit.Blocker
				if !errors.As(err, &blocker) || blocker.Kind != test.kind ||
					errors.Is(err, gosplit.ErrBlocked) == (test.kind == gosplit.DirectoryUnobserved) {
					t.Fatalf("typed original cause was lost: %v", err)
				}
			}
			if result.SourceSHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(test.source))) ||
				result.Requirements[0].State != test.first || result.Requirements[1].State != test.next ||
				result.Requirements[2].State != PrerequisiteUnobserved || result.ReadyForProjection() != (err == nil) {
				t.Fatalf("preparation was confused with package conformance: %+v", result)
			}
			assertGoSplitRequirementOrigins(t, result)
			if result.PolicySource.SHA256 != fmt.Sprintf("%x", sha256.Sum256(goSplitSelectionSource)) ||
				result.PolicySource.ByteLength != len(goSplitSelectionSource) {
				t.Fatal("next-step policy lost its own source-byte identity")
			}
			next := result.NextStep()
			if next.F_next_operation != test.operation {
				t.Fatalf("source-owned next operation = %q; want %q", next.F_next_operation, test.operation)
			}
			linked := false
			for _, requirement := range result.Requirements {
				linked = linked || requirement.Origin.ID == next.F_next_requirement
			}
			if !linked {
				t.Fatal("source decision points outside the admitted prerequisite chain")
			}
			after, _ := json.Marshal(report)
			if string(before) != string(after) {
				t.Fatal("inspection changed raw debt or directory observations")
			}
			assertSplitPlannerSource(t, root, test.source)
		})
	}
}

func TestGoSplitPreparationPreservesUnreadableAndMalformedInputs(t *testing.T) {
	root, report := splitPlannerFixture(t, "package fixture\nfunc Broken(\n")
	for _, subject := range []string{"missing.go", "fixture.go"} {
		result, err := InspectGoSplit(root, subject, report)
		var blocker *gosplit.Blocker
		if err == nil || errors.As(err, &blocker) || result.SourceSHA256 != "" ||
			result.Requirements[0].State != PrerequisiteUnknown || result.ReadyForProjection() || result.Plan != nil {
			t.Fatalf("missing/parser input became a capacity result: %+v, %v", result, err)
		}
	}
}

func assertGoSplitRequirementOrigins(t *testing.T, result GoSplitPreparation) {
	t.Helper()
	raw := operationContractFixture(t)
	for index, suffix := range []string{"rendered-capacity", "directory-capacity", "projected-conformance"} {
		step := result.Requirements[index]
		if step.Origin.ID != "gooo://meta-operation-input-contract/activity/go-split-"+suffix ||
			!step.Origin.Span.Valid() || step.Origin.Span.File != "operation-input-contract.gooo" ||
			!strings.Contains(string(raw[step.Origin.Span.Start:step.Origin.Span.End]), step.Origin.ID) {
			t.Fatalf("requirement lost its actual authored source: %+v", step)
		}
	}
}

func TestGoSplitPreparationRereadsChangedSource(t *testing.T) {
	root, report := splitPlannerFixture(t, splitPlannerSource)
	first, err := InspectGoSplit(root, "fixture.go", report)
	if err != nil {
		t.Fatal(err)
	}
	changed := "//다른 기원\n" + splitPlannerSource
	if err := os.WriteFile(filepath.Join(root, "fixture.go"), []byte(changed), 0o640); err != nil {
		t.Fatal(err)
	}
	second, err := InspectGoSplit(root, "fixture.go", report)
	if err != nil || first.SourceSHA256 == second.SourceSHA256 || !second.ReadyForProjection() ||
		first.Contract.SourceDigest != second.Contract.SourceDigest {
		t.Fatalf("file identity and contract identity were conflated: %v", err)
	}
	assertSplitPlannerSource(t, root, changed)
}

func TestGoSplitSelectionKeepsUnknownFactsUnready(t *testing.T) {
	for _, input := range []R_SplitObservations{
		{},
		{F_rendered_state: "MAYBE", F_directory_state: "OBSERVED"},
		{F_rendered_state: "OBSERVED", F_directory_state: "MAYBE"},
		{F_rendered_state: "OBSERVED", F_directory_state: "OBSERVED", F_native_blocker: "UNKNOWN_BLOCKER"},
	} {
		if output := Evaluate(input); output.F_ready_for_projection {
			t.Fatalf("source policy promoted unknown native facts: %+v", input)
		}
	}
	var preparation GoSplitPreparation
	preparation.Requirements[0].State, preparation.Requirements[1].State = PrerequisiteObserved, PrerequisiteObserved
	if preparation.ReadyForProjection() {
		t.Fatal("labels alone supplied an absent native plan")
	}
}
