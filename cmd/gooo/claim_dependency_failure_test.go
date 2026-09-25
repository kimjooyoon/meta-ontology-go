package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

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
