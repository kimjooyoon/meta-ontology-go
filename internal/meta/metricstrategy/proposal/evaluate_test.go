package proposal

import "testing"

func TestConformancePlanProvesThreeExactCoordinates(t *testing.T) {
	coordinates, facts, err := generationCoordinates()
	if err != nil { t.Fatal(err) }
	if len(coordinates) != 3 || facts.Decision != "PLAN" || facts.Actions != 2 || facts.Promotion || facts.Writes {
		t.Fatalf("unexpected proposal facts: %+v %+v", coordinates, facts)
	}
	for _, coordinate := range coordinates {
		if coordinate.Status != "SATISFIED" || coordinate.EvidenceDigest == "" { t.Fatalf("unproven coordinate: %+v", coordinate) }
	}
}

func TestFixedRegistryAndUnknownResolution(t *testing.T) {
	if len(Registry()) != 8 { t.Fatalf("coordinate denominator changed: %d", len(Registry())) }
	coordinates := make([]Coordinate, len(registry))
	for index, spec := range registry { coordinates[index] = Coordinate{CoordinateSpec: spec, Status: "SATISFIED", EvidenceDigest: "sha256:evidence"} }
	coordinates[3].Status = "UNRESOLVED"
	summary := summarize(coordinates)
	decision, reason := decisionFor(summary)
	if summary.Satisfied != 7 || summary.Total != 8 || summary.Unresolved != 1 || summary.ReadinessBPS != 8750 || decision != "FAIL_CLOSED" || reason != "CHANGE_PROPOSAL_EVIDENCE_UNKNOWN" {
		t.Fatalf("unknown evidence did not lower resolution: %+v %s/%s", summary, decision, reason)
	}
}
