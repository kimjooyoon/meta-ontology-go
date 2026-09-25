package workfrontier

import (
	"bytes"
	"embed"
	"encoding/json"
	"testing"
)

//go:embed testdata/r4_cases.json
var r4FixtureData embed.FS

type r4Fixture struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func TestR4StrictFixtures(t *testing.T) {
	var fixtures []r4Fixture
	data, err := r4FixtureData.ReadFile("testdata/r4_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			input := r4FixtureInput(t, fixture.Name)
			got := EvaluateR4(input)
			if got.Status != fixture.Status || got.Reason != fixture.Reason {
				t.Fatalf("result = %s/%s, want %s/%s", got.Status, got.Reason, fixture.Status, fixture.Reason)
			}
			if (got.Status == R4StatusUnknown || got.Status == R4StatusFailClosed) &&
				len(got.SelectedIDs) != 0 {
				t.Fatalf("non-pass result selected %v", got.SelectedIDs)
			}
			if got.GraphDigest == "" || got.SCCDigest == "" || got.CondensationDigest == "" || got.RuleDigest == "" {
				t.Fatal("missing canonical graph or rule digest")
			}
			encoded, err := EncodeR4ResultJSON(got)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(encoded, []byte("proof_valid")) || bytes.Contains(encoded, []byte("promotion_authorized")) {
				t.Fatalf("result emitted forbidden authorization field: %s", encoded)
			}
		})
	}
}
