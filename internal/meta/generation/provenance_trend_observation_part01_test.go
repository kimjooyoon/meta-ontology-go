package generation

import "testing"

func TestObserveProvenanceTrendPreservesDirectionAndEvidence(t *testing.T) {
	previous := envelopeDigestString("previous-evidence")
	current := envelopeDigestString("current-evidence")
	cases := []struct {
		name      string
		previous  int
		current   int
		direction string
	}{
		{name: "improved", previous: 2, current: 3, direction: ProvenanceTrendImproved},
		{name: "regressed", previous: 3, current: 2, direction: ProvenanceTrendRegressed},
		{name: "stable", previous: 3, current: 3, direction: ProvenanceTrendStable},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			observation := ObserveProvenanceTrend(previous, current, testCase.previous, testCase.current)
			if observation.Direction != testCase.direction {
				t.Fatalf("direction = %q, want %q", observation.Direction, testCase.direction)
			}
			if err := observation.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestObserveProvenanceTrendPreservesUnknownAndMalformedEvidence(t *testing.T) {
	unknown := ObserveProvenanceTrend("", "", 0, 1)
	if unknown.Direction != ProvenanceTrendUnknown || unknown.Reason != "PROVENANCE_TREND_UNKNOWN" {
		t.Fatalf("unknown = %#v", unknown)
	}
	if err := unknown.Validate(); err != nil {
		t.Fatalf("unknown Validate() error = %v", err)
	}

	malformed := ObserveProvenanceTrend("not-a-digest", "also-not-a-digest", 0, 1)
	if malformed.Direction != ProvenanceTrendUnknown ||
		malformed.Reason != "MALFORMED_PROVENANCE_EVIDENCE_DIGEST" {
		t.Fatalf("malformed = %#v", malformed)
	}
	if err := malformed.Validate(); err != nil {
		t.Fatalf("malformed Validate() error = %v", err)
	}
}
