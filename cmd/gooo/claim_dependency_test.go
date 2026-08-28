package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestClaimDependencyExampleReconstructsTypedEdges(t *testing.T) {
	source, err := os.ReadFile("testdata/claim-dependency/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(source, []byte(claimDependencyCandidateID)) {
		t.Fatal("candidate id is not directly mapped in Gooo source")
	}
	file, diagnostics := syntax.ParseFile("main.gooo", string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics.Error())
	}
	report := resolveClaimDependencies("main.gooo", source, file)
	if report.Decision != claimDependencyObserved || report.Resolution.State != claimStateClosed || report.Summary.ActivitiesObserved != 6 || report.Summary.RecoverableRoots != 1 || report.Summary.TypedDeclarations != 5 || report.Summary.DependencyInputs != 8 || report.Summary.TypedEdges != 8 || report.Summary.EdgeKindsObserved != 4 || report.Summary.UnresolvedInputs != 0 || report.Summary.CyclicActivities != 0 {
		t.Fatalf("claim dependency denominator changed: %#v", report)
	}
	if report.KindCounts.Requires != 3 || report.KindCounts.Supports != 2 || report.KindCounts.Contradicts != 2 || report.KindCounts.FailureEntailment != 1 || len(report.Indicators) != 8 {
		t.Fatalf("typed edge reconstruction changed: %#v", report)
	}
	declared := make(map[string]bool)
	for _, node := range report.Nodes {
		declared[node.Activity] = true
	}
	for _, indicator := range report.Indicators {
		for _, activity := range indicator.Activities {
			if !declared[activity] {
				t.Fatalf("indicator %q detached from Gooo activity %q", indicator.ID, activity)
			}
		}
	}
}

func TestClaimDependencyMissingProducerLowersResolution(t *testing.T) {
	source := []byte("package claims\nnamespace claims\nentity External id \"gooo://claims/external\"\nentity Result id \"gooo://claims/result\"\nactivity Derived(External) -> Result computes \"claim.edge:requires|external-evidence\"\n")
	file, diagnostics := syntax.ParseFile("missing.gooo", string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics.Error())
	}
	report := resolveClaimDependencies("missing.gooo", source, file)
	if report.Decision != claimDependencyIncomplete || report.Resolution.State != claimStateUnknown || report.Resolution.Stage == nil || *report.Resolution.Stage != "DEPENDENCY_DISCOVERY" || report.Resolution.Step == nil || *report.Resolution.Step != "BIND_INPUT_PRODUCER" || report.Resolution.Reason != "CLAIM_INPUT_PRODUCER_UNAVAILABLE" || report.Resolution.UnknownClass == nil || *report.Resolution.UnknownClass != "DIRECT_MISSING" || report.Resolution.NextOperation != "DECLARE_INPUT_PRODUCER" || report.Summary.UnresolvedInputs != 1 {
		t.Fatalf("missing producer resolution changed: %#v", report)
	}
}

func TestClaimDependencyRejectsUnsupportedKindAndCycle(t *testing.T) {
	invalid := []byte("package claims\nnamespace claims\nentity Seed id \"gooo://claims/seed\"\nentity RootState id \"gooo://claims/root\"\nentity Result id \"gooo://claims/result\"\nactivity Root(Seed) -> RootState computes \"claim.observe:recoverable|root\"\nactivity Derived(RootState) -> Result computes \"claim.edge:implies|derived\"\n")
	cycle := []byte("package claims\nnamespace claims\nentity AState id \"gooo://claims/a\"\nentity BState id \"gooo://claims/b\"\nactivity A(BState) -> AState computes \"claim.edge:requires|a\"\nactivity B(AState) -> BState computes \"claim.edge:requires|b\"\n")
	for _, tc := range []struct {
		name, reason string
		source       []byte
	}{
		{name: "unsupported-kind", source: invalid, reason: "CLAIM_DEPENDENCY_EDGE_KIND_UNSUPPORTED"},
		{name: "cycle", source: cycle, reason: "CLAIM_DEPENDENCY_CYCLE_DETECTED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file, diagnostics := syntax.ParseFile(tc.name+".gooo", string(tc.source))
			if diagnostics.HasErrors() {
				t.Fatal(diagnostics.Error())
			}
			report := resolveClaimDependencies(tc.name+".gooo", tc.source, file)
			if report.Decision != claimDependencyFailed || report.Resolution.State != claimStateRefuted || report.Resolution.Reason != tc.reason {
				t.Fatalf("refutation changed: %#v", report)
			}
		})
	}
}

func TestClaimDependenciesCommandEmitsClosedAndUnknownReceipts(t *testing.T) {
	valid, err := os.ReadFile("testdata/claim-dependency/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	missing := "package claims\nnamespace claims\nentity External id \"gooo://claims/external\"\nentity Result id \"gooo://claims/result\"\nactivity Derived(External) -> Result computes \"claim.edge:requires|external-evidence\"\n"
	for _, tc := range []struct {
		name, source string
		code         int
		decision     string
	}{
		{name: "closed", source: string(valid), code: exitOK, decision: claimDependencyObserved},
		{name: "unknown", source: missing, code: exitFailure, decision: claimDependencyIncomplete},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runClaim([]string{"dependencies", "main.gooo", "--json"}, fixtureReader{source: tc.source}, SyntaxSourceParser{}, &stdout, &stderr)
			var report claimDependencyReport
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatal(err)
			}
			if code != tc.code || report.Decision != tc.decision {
				t.Fatalf("code=%d report=%#v stderr=%q", code, report, stderr.String())
			}
		})
	}
}
