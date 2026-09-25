package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

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
