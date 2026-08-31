package languagereadiness

import "testing"

func TestUseCaseCurrentCatalogIsSevenOfTwentyFour(t *testing.T) {
	snapshot, err := Evaluate(artifactFixture("PASS", currentConceptIDs...))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Decision != "PASS" {
		t.Fatalf("decision = %q", snapshot.Decision)
	}
	if snapshot.Summary.Completed != 7 || snapshot.Summary.Total != 24 {
		t.Fatalf("completion = %d/%d", snapshot.Summary.Completed, snapshot.Summary.Total)
	}
	if snapshot.Summary.ReadinessBPS != 2916 || snapshot.Summary.NotSatisfied != 17 {
		t.Fatalf("summary = %+v", snapshot.Summary)
	}
	if snapshot.Summary.RatioNumerator != 7 || snapshot.Summary.RatioDenominator != 24 {
		t.Fatalf("ratio = %d/%d", snapshot.Summary.RatioNumerator, snapshot.Summary.RatioDenominator)
	}
}

func TestUseCaseUnknownEvidenceLowersResolution(t *testing.T) {
	snapshot, err := Evaluate(artifactFixture("FUTURE_DECISION", currentConceptIDs...))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Decision != "LOWER_RESOLUTION" || snapshot.Summary.Unresolved != 24 {
		t.Fatalf("unexpected resolution: decision=%s summary=%+v", snapshot.Decision, snapshot.Summary)
	}
	if snapshot.Summary.Completed != 0 {
		t.Fatalf("unknown evidence completed %d obligations", snapshot.Summary.Completed)
	}
}

func TestUseCaseUnregisteredClaimsDoNotChangeTheCount(t *testing.T) {
	ids := append(append([]string(nil), currentConceptIDs...), "qualitative-language-progress")
	snapshot, err := Evaluate(artifactFixture("PASS", ids...))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Summary.Completed != 7 {
		t.Fatalf("unregistered claim changed completion to %d", snapshot.Summary.Completed)
	}
	replayed, err := Evaluate(artifactFixture("PASS", ids...))
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Digest != snapshot.Digest {
		t.Fatalf("replay digest mismatch: %s != %s", replayed.Digest, snapshot.Digest)
	}
}
