package bidir

import (
	"reflect"
	"strings"
	"testing"
)

func TestBXEvidenceRejectsTamperedCanonicalDeltaEvidence(t *testing.T) {
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*BXDeltaEvidence){
		"canonical-json": func(delta *BXDeltaEvidence) {
			delta.CanonicalJSON = strings.Replace(delta.CanonicalJSON, `"partial_observation":false`, `"partial_observation":true`, 1)
		},
		"locality-json": func(delta *BXDeltaEvidence) {
			delta.LocalityCanonicalJSON = "{}"
		},
		"port-order-hash": func(delta *BXDeltaEvidence) {
			delta.PortOrderHash = digest("tampered port sequence")
		},
		"order-hash": func(delta *BXDeltaEvidence) {
			tampered := digest("tampered order")
			delta.CanonicalJSON = strings.Replace(delta.CanonicalJSON, `"order_hash":"`+delta.OrderHash+`"`, `"order_hash":"`+tampered+`"`, 1)
			delta.OrderHash = tampered
		},
		"evidence-hash": func(delta *BXDeltaEvidence) {
			delta.EvidenceHash = digest("tampered evidence")
			delta.EvidenceSpans.Hash = delta.EvidenceHash
		},
		"sequence-hash-format": func(delta *BXDeltaEvidence) {
			delta.SequenceHash = "not-a-sha256"
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			copyEvidence := evidence
			mutate(&copyEvidence.Delta)
			if err := copyEvidence.validate(); err == nil {
				t.Fatal("tampered canonical delta evidence was accepted")
			}
		})
	}
}
func TestEvidenceRetainsSameFactKeyWithDistinctIDsAndSpans(t *testing.T) {
	first := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "a.go", Start: 1, End: 2})
	first.EvidenceID = "urn:gooo:evidence:explicit-a"
	second := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "b.go", Start: 3, End: 4})
	second.EvidenceID = "urn:gooo:evidence:explicit-b"
	evidence := evidenceSpans(FactSet{first, second})
	if evidence.IDCount != 2 || evidence.SpanCount != 2 || evidence.IDs[0] == evidence.IDs[1] {
		t.Fatalf("same FactKey evidence was collapsed: %#v", evidence)
	}
	if evidence.EvidenceIDAuthority != "explicit" || !reflect.DeepEqual(evidence.IDs, []string{first.EvidenceID, second.EvidenceID}) {
		t.Fatalf("explicit evidence IDs were not retained: %#v", evidence)
	}
	if evidence.FactKeys[0] != evidence.FactKeys[1] || evidence.Hash == "" {
		t.Fatalf("same-edge evidence boundary lost key/hash: %#v", evidence)
	}
}
