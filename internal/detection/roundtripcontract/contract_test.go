package roundtripcontract

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
)

func TestMinimalFixtureMeasurement(t *testing.T) {
	fixture := MinimalFixture()
	if fixture.Measurement.Nodes != 3 || fixture.Measurement.Facts != 2 || fixture.Measurement.Regions != 3 {
		t.Fatalf("unexpected fixture measurement: %#v", fixture.Measurement)
	}
	if fixture.Measurement.SourceBytes != int64(len(fixture.DSL)+len(fixture.Go)) || fixture.Measurement.SourceBytes == 0 {
		t.Fatalf("source byte measurement is not reproducible: %#v", fixture.Measurement)
	}
	if len(fixture.Artifacts) != 4 {
		t.Fatalf("expected four stage artifacts, got %d", len(fixture.Artifacts))
	}
}

func TestScenariosClassifyPassFailAndDeferred(t *testing.T) {
	seen := make(map[Outcome]bool)
	for _, scenario := range MinimalScenarios() {
		assessment, err := Assess(scenario)
		if err != nil {
			t.Fatalf("%s: %v", scenario.CaseID, err)
		}
		if !assessment.Accepted {
			t.Fatalf("scenario did not match expected outcome: %#v", assessment)
		}
		seen[assessment.Actual] = true
		if assessment.Actual != OutcomePass && assessment.MergeEligible {
			t.Fatalf("non-pass outcome became merge-eligible: %#v", assessment)
		}
	}
	for _, outcome := range []Outcome{OutcomePass, OutcomeFail, OutcomeDeferred} {
		if !seen[outcome] {
			t.Fatalf("missing scenario outcome %q", outcome)
		}
	}
}

func TestNegativeCaseWithoutFindingIsNotAccepted(t *testing.T) {
	scenario := MinimalScenarios()[1]
	scenario.Evidence.Findings = nil
	scenario.Evidence.Outcome = OutcomePass
	assessment, err := Assess(scenario)
	if err != nil {
		t.Fatal(err)
	}
	if assessment.Accepted || assessment.MergeEligible {
		t.Fatalf("semantic counterexample was accepted without a finding: %#v", assessment)
	}
}

func TestDeferredStageCannotMasqueradeAsPass(t *testing.T) {
	scenario := MinimalScenarios()[2]
	scenario.Expected = OutcomePass
	assessment, err := Assess(scenario)
	if err != nil {
		t.Fatal(err)
	}
	if assessment.Accepted || assessment.MergeEligible || assessment.Actual != OutcomeDeferred {
		t.Fatalf("deferred stage was treated as success: %#v", assessment)
	}
	for _, artifact := range scenario.Evidence.Artifacts {
		if artifact.Stage == StageLiftedIR {
			t.Fatal("deferred evidence claimed an unimplemented lifted IR artifact")
		}
	}
}

func TestCanonicalJSONRoundTrip(t *testing.T) {
	evidence := MinimalScenarios()[1].Evidence
	encoded, err := CanonicalJSON(evidence)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(encoded, []byte{'\n'}) || !strings.Contains(string(encoded), `"outcome":"fail"`) {
		t.Fatalf("evidence JSON is not CI-friendly: %s", encoded)
	}
	var decoded Evidence
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	reencoded, err := CanonicalJSON(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatalf("evidence JSON was not canonical:\n%s\n%s", encoded, reencoded)
	}
}

func TestCanonicalJSONOrderProperty(t *testing.T) {
	base := MinimalScenarios()[1].Evidence
	base.Findings = append(base.Findings,
		Finding{Rule: "round-trip", Path: "dsl", Identity: "billing://entity/payment", Detail: "fact was removed"},
		Finding{Rule: "locality", Path: "generated-go", Identity: "billing://entity/unrelated", Detail: "unrelated region changed"},
	)
	property := func(seed uint64) bool {
		candidate := base.Normalize()
		random := rand.New(rand.NewSource(int64(seed)))
		random.Shuffle(len(candidate.Findings), func(i, j int) {
			candidate.Findings[i], candidate.Findings[j] = candidate.Findings[j], candidate.Findings[i]
		})
		left, leftErr := CanonicalJSON(base)
		right, rightErr := CanonicalJSON(candidate)
		return leftErr == nil && rightErr == nil && bytes.Equal(left, right)
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 128}); err != nil {
		t.Fatal(err)
	}
}

func TestEvidenceRejectsUnsupportedSuccessClaims(t *testing.T) {
	evidence := MinimalScenarios()[0].Evidence
	evidence.Outcome = OutcomeFail
	if err := evidence.Validate(); err == nil {
		t.Fatal("inconsistent outcome was accepted")
	}
	evidence = MinimalScenarios()[2].Evidence
	evidence.DeferredReason = ""
	if err := evidence.Validate(); err == nil {
		t.Fatal("deferred evidence without a reason was accepted")
	}
}
