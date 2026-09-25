package bidir

import (
	"reflect"
	"strings"
	"testing"
)

func TestEvidenceRecordsPreserveDuplicateFactKeysByPermutation(t *testing.T) {
	first := duplicateEvidenceFact("evidence-b", 30)
	second := duplicateEvidenceFact("evidence-a", 10)
	left := evidenceSpans(FactSet{first, second})
	right := evidenceSpans(FactSet{second, first})
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("permuted duplicate evidence is not deterministic:\nleft %#v\nright %#v", left, right)
	}
	if left.EvidenceIDAuthority != "explicit" || len(left.Records) != 2 || left.Records[0].EvidenceID != second.EvidenceID || !left.Records[0].HasSpan {
		t.Fatalf("duplicate evidence records lost order or authority: %#v", left)
	}
	if left.Records[0].FactKey != left.Records[1].FactKey || left.IDCount != 2 || left.SpanCount != 2 {
		t.Fatalf("same-edge cardinality was not retained: %#v", left)
	}
}
func TestEvidenceAuthorityDistinguishesExplicitAndDerivedRecords(t *testing.T) {
	explicit := evidenceSpans(FactSet{duplicateEvidenceFact("explicit", 1)})
	derivedFact := duplicateEvidenceFact("", 1)
	derived := evidenceSpans(FactSet{derivedFact})
	mixed := evidenceSpans(FactSet{duplicateEvidenceFact("explicit", 1), derivedFact})
	if explicit.EvidenceIDAuthority != "explicit" || derived.EvidenceIDAuthority != "derived-non-authoritative" || mixed.EvidenceIDAuthority != "mixed-non-authoritative" {
		t.Fatalf("evidence authority classification is not fail-closed: explicit=%q derived=%q mixed=%q", explicit.EvidenceIDAuthority, derived.EvidenceIDAuthority, mixed.EvidenceIDAuthority)
	}
}
func TestBXDeltaCanonicalExportsEvidenceRecordsAndCounts(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	changes := FactDelta{Added: FactSet{duplicateEvidenceFact("evidence-a", 10), duplicateEvidenceFact("evidence-b", 20)}}
	evidence := makeDeltaEvidenceUnchecked(changes, Locality{}, false, base, base)
	if err := validateDeltaEvidence(evidence); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"evidence_records", "evidence_id_count", "evidence_span_count", "evidence_id_authority"} {
		if !strings.Contains(evidence.CanonicalJSON, `"`+field+`"`) {
			t.Fatalf("canonical delta omitted %q: %s", field, evidence.CanonicalJSON)
		}
	}
}
