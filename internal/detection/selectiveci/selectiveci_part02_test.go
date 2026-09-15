package selectiveci

import (
	"testing"
)

func TestCanonicalPermutationIsByteIdentical(t *testing.T) {
	left := completeInput()
	right := completeInput()
	right.Base.Files[0].SemanticIDs = []string{"urn:selectiveci:entity/order"}
	right.Head.Files[0].SemanticIDs = []string{"urn:selectiveci:entity/order"}
	right.Receipts[0], right.Receipts[1] = right.Receipts[1], right.Receipts[0]
	right.ProvenancePaths[0], right.ProvenancePaths[1] = right.ProvenancePaths[1], right.ProvenancePaths[0]
	leftBytes, err := EncodeJSON(left)
	if err != nil {
		t.Fatal(err)
	}
	rightBytes, err := EncodeJSON(right)
	if err != nil {
		t.Fatal(err)
	}
	if string(leftBytes) != string(rightBytes) {
		t.Fatalf("permutation changed canonical input:\n%s\n%s", leftBytes, rightBytes)
	}
	if Plan(left).Canonical() != Plan(right).Canonical() {
		t.Fatal("permutation changed canonical output")
	}
}
func TestStrictJSONRejectsDuplicateAndUnknownFields(t *testing.T) {
	if _, err := DecodeJSON([]byte(`{"schema_version":"gooo/selective-ci/v1","schema_version":"gooo/selective-ci/v1"}`)); err == nil {
		t.Fatal("duplicate field was accepted")
	}
	if _, err := DecodeJSON([]byte(`{"schema_version":"gooo/selective-ci/v1","unknown":1}`)); err == nil {
		t.Fatal("unknown field was accepted")
	}
}
func TestStrictJSONRoundTripAndSealedPlan(t *testing.T) {
	input := completeInput()
	encoded, err := EncodeJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if input.Canonical() != decoded.Canonical() {
		t.Fatalf("round trip changed canonical input")
	}
	planBytes, err := EncodePlanJSON(Plan(input))
	if err != nil {
		t.Fatal(err)
	}
	plan := PlanJSON(planBytes)
	if plan.Status != StatusFullSuiteFallback || plan.ReasonCode != ReasonInvalidInput {
		t.Fatalf("sealed plan should not be accepted as input: %#v", plan)
	}
	if !containsString(string(planBytes), `"canonical_digest":"`) {
		t.Fatalf("plan output omitted canonical digest: %s", planBytes)
	}
}
func FuzzPlanJSONNeverPanics(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(t *testing.T, data []byte) {
		result := PlanJSON(data)
		if result.CanonicalDigest == "" || result.CanonicalDigest != result.StableDigest() {
			t.Fatalf("unsealed result for arbitrary input: %#v", result)
		}
		if result.Status != StatusFullSuiteFallback && result.Status != StatusSelective {
			t.Fatalf("unknown result status %q", result.Status)
		}
	})
}
