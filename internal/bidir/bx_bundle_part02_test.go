package bidir

import (
	_ "embed"
	"encoding/json"
	"testing"
)

type bxBundleTransaction struct {
	Observer string        `json:"observer"`
	Observed bool          `json:"observed"`
	Atomic   bool          `json:"atomic"`
	NoWrite  bool          `json:"no_write"`
	Deferred bool          `json:"deferred"`
	Before   bxBundleState `json:"before"`
	After    bxBundleState `json:"after"`
}
type bxBundleState struct {
	Semantic string `json:"semantic"`
	Source   string `json:"source"`
	Region   string `json:"region"`
	Slot     string `json:"slot"`
	Bytes    string `json:"bytes"`
	LStat    string `json:"lstat"`
}
type bxBundlePolicy struct {
	PartialNoDelete       bool     `json:"partial_no_delete"`
	CandidatePromotion    bool     `json:"candidate_promotion"`
	AuthoritativeEvidence string   `json:"authoritative_evidence_ids"`
	ProvenanceMapping     string   `json:"provenance_mapping"`
	Deferred              []string `json:"deferred"`
}

func TestBXEvidenceBundleIsNonEmptyAndBounded(t *testing.T) {
	var bundle bxBundle
	if err := json.Unmarshal(bxBillingEvidenceBundle, &bundle); err != nil {
		t.Fatal(err)
	}
	if bundle.SchemaVersion != BXEvidenceSchemaVersion || bundle.BundleKind != "bx-observer-provenance" || bundle.Status != "bounded-contract-only" || bundle.FeatureGreen {
		t.Fatalf("bundle is not bounded contract evidence: %#v", bundle)
	}
	if bundle.Source.IntegrationBase != "ddaf3732d8539f95311c061ef33971414f9dbbeb" || bundle.Source.ParentHead != "a74ee2234b5fde1d6bec20ae043f9b65ee8b1dfa" || bundle.Source.Fixture != "billing" {
		t.Fatalf("bundle source tuple is incomplete: %#v", bundle.Source)
	}
	if len(bundle.Base) != 6 || len(bundle.Deltas) != 2 || len(bundle.Transactions) != 2 {
		t.Fatalf("bundle inventory is incomplete: %#v", bundle)
	}
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	assertBundleBase(t, bundle.Base, evidence.Base)
	if !bundle.RoundTrip.GetPut || !bundle.RoundTrip.PutGet || !bundle.RoundTrip.SemanticEquivalent {
		t.Fatalf("bundle omitted round-trip evidence: %#v", bundle.RoundTrip)
	}
	assertBundleDelta(t, bundle.Deltas["accepted"], evidence.Delta)
	assertBundleDelta(t, bundle.Deltas["partial"], evidence.PartialDelta)
	assertBundleTransaction(t, bundle.Transactions["rejected"], evidence.RejectedTransaction)
	if bundle.Policy.CandidatePromotion || !bundle.Policy.PartialNoDelete || bundle.Policy.AuthoritativeEvidence != "deferred" || bundle.Policy.ProvenanceMapping != "deferred" {
		t.Fatalf("bundle policy is not fail-closed: %#v", bundle.Policy)
	}
	for _, seam := range deferredBXSeams() {
		if !containsString(bundle.Policy.Deferred, seam) {
			t.Fatalf("bundle omitted deferred seam %q", seam)
		}
	}
}
