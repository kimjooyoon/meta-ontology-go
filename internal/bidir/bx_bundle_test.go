package bidir

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
)

//go:embed testdata/bx_billing_evidence_bundle.json
var bxBillingEvidenceBundle []byte

type bxBundleArtifact struct {
	Hash  string `json:"hash"`
	Count int    `json:"count"`
}

type bxBundle struct {
	SchemaVersion string                         `json:"schema_version"`
	BundleKind    string                         `json:"bundle_kind"`
	Status        string                         `json:"status"`
	FeatureGreen  bool                           `json:"feature_green"`
	Base          map[string]bxBundleArtifact    `json:"base"`
	RoundTrip     bxBundleRoundTrip              `json:"round_trip"`
	Source        bxBundleSource                 `json:"source"`
	Deltas        map[string]bxBundleDelta       `json:"deltas"`
	Transactions  map[string]bxBundleTransaction `json:"transactions"`
	Policy        bxBundlePolicy                 `json:"policy"`
}

type bxBundleSource struct {
	IntegrationBase string `json:"integration_base"`
	ParentHead      string `json:"parent_head"`
	Fixture         string `json:"fixture"`
}

type bxBundleRoundTrip struct {
	GetPut             bool `json:"get_put"`
	PutGet             bool `json:"put_get"`
	SemanticEquivalent bool `json:"semantic_equivalence"`
}

type bxBundleDelta struct {
	SequenceHash       string           `json:"sequence_hash"`
	OrderHash          string           `json:"order_hash"`
	PortOrderHash      string           `json:"port_order_hash"`
	RelationOrderHash  string           `json:"relation_order_hash"`
	PortSequence       []string         `json:"port_sequence"`
	RelationSequence   []string         `json:"relation_sequence"`
	Candidates         []string         `json:"candidates"`
	Closure            bxBundleClosure  `json:"closure"`
	Evidence           bxBundleEvidence `json:"evidence"`
	PartialObservation bool             `json:"partial_observation"`
}

type bxBundleClosure struct {
	Touched  []string `json:"touched"`
	Affected []string `json:"affected"`
	Members  []string `json:"members"`
	Hash     string   `json:"hash"`
}

type bxBundleEvidence struct {
	IDs                 []string                 `json:"ids"`
	FactKeys            []string                 `json:"fact_keys"`
	Spans               []string                 `json:"spans"`
	Records             []bxBundleEvidenceRecord `json:"records"`
	IDCount             int                      `json:"id_count"`
	SpanCount           int                      `json:"span_count"`
	Hash                string                   `json:"hash"`
	EvidenceIDAuthority string                   `json:"evidence_id_authority"`
}

type bxBundleEvidenceRecord struct {
	EvidenceID string `json:"evidence_id"`
	FactKey    string `json:"fact_key"`
	Span       string `json:"span"`
	HasSpan    bool   `json:"has_span"`
}

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

func assertBundleBase(t *testing.T, got map[string]bxBundleArtifact, want BXBaseEvidence) {
	t.Helper()
	wantArtifacts := map[string]BXArtifactEvidence{"dsl": want.DSL, "ir": want.IR, "go": want.Go, "source_map": want.SourceMap, "evidence": want.Evidence, "provenance": want.Provenance}
	for name, artifact := range wantArtifacts {
		if got[name].Hash != artifact.Hash || got[name].Count != artifact.Count {
			t.Fatalf("bundle base %s mismatch: got=%#v want=%#v", name, got[name], artifact)
		}
	}
}

func assertBundleDelta(t *testing.T, got bxBundleDelta, want BXDeltaEvidence) {
	t.Helper()
	if got.SequenceHash != want.SequenceHash || got.OrderHash != want.OrderHash || got.PortOrderHash != want.PortOrderHash || got.RelationOrderHash != want.RelationOrderHash || got.PartialObservation != want.PartialObservation {
		t.Fatalf("bundle delta hashes mismatch: got=%#v want=%#v", got, want)
	}
	if !reflect.DeepEqual(got.PortSequence, want.PortSequence) || !reflect.DeepEqual(got.RelationSequence, want.RelationSequence) || !reflect.DeepEqual(got.Candidates, want.Candidates) {
		t.Fatal("bundle delta sequence or candidate evidence mismatch")
	}
	if !reflect.DeepEqual(got.Closure.Touched, idsAsStrings(want.Locality.Touched)) || !reflect.DeepEqual(got.Closure.Affected, idsAsStrings(want.Locality.Affected)) || !reflect.DeepEqual(got.Closure.Members, idsAsStrings(want.ClosureMembers)) || got.Closure.Hash != want.LocalityClosureHash {
		t.Fatal("bundle locality closure evidence mismatch")
	}
	if !reflect.DeepEqual(got.Evidence.IDs, want.EvidenceSpans.IDs) || !reflect.DeepEqual(got.Evidence.FactKeys, want.EvidenceSpans.FactKeys) || !reflect.DeepEqual(got.Evidence.Records, bundleRecords(want.EvidenceSpans.Records)) || got.Evidence.IDCount != want.EvidenceSpans.IDCount || got.Evidence.SpanCount != want.EvidenceSpans.SpanCount || got.Evidence.Hash != want.EvidenceHash || got.Evidence.EvidenceIDAuthority != want.EvidenceSpans.EvidenceIDAuthority {
		t.Fatalf("bundle evidence ID/span set mismatch: got=%#v want=%#v", got.Evidence, want.EvidenceSpans)
	}
}

func bundleRecords(records []BXEvidenceRecord) []bxBundleEvidenceRecord {
	values := make([]bxBundleEvidenceRecord, len(records))
	for index, record := range records {
		values[index] = bxBundleEvidenceRecord{EvidenceID: record.EvidenceID, FactKey: record.FactKey, Span: spanText(record.Span), HasSpan: record.HasSpan}
	}
	return values
}

func assertBundleTransaction(t *testing.T, got bxBundleTransaction, want BXTransactionEvidence) {
	t.Helper()
	if got.Observer != want.ObserverKind || got.Observed != want.Observed || got.Atomic != want.Atomic || got.NoWrite != want.NoWrite || got.Deferred != want.Deferred || got.Before != stateBundle(want.Before) || got.After != stateBundle(want.After) {
		t.Fatalf("bundle transaction mismatch: got=%#v want=%#v", got, want)
	}
}

func stateBundle(state BXStateEvidence) bxBundleState {
	return bxBundleState{Semantic: state.Semantic, Source: state.Source, Region: state.Region, Slot: state.Slot, Bytes: state.Bytes, LStat: state.LStat}
}

func idsAsStrings(ids []ID) []string {
	values := make([]string, len(ids))
	for index, id := range ids {
		values[index] = string(id)
	}
	return values
}
