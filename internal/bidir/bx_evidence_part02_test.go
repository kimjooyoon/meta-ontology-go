package bidir

import (
	"reflect"
	"strings"
	"testing"
)

func TestBXEvidenceMatchesBillingGolden(t *testing.T) {
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{
		"schema_version=" + BXEvidenceSchemaVersion,
		"fixture=billing",
		"get_put=pass",
		"put_get=pass",
		"semantic_equivalence=pass",
		"accepted_sequence_hash=",
		"accepted_order_hash=",
		"accepted_locality_closure_hash=",
		"accepted_port_order_hash=",
		"accepted_relation_order_hash=",
		"accepted_evidence_id_count=1",
		"accepted_evidence_span_count=1",
		"accepted_before=",
		"accepted_after=",
		"rejected_observer=memory-source",
		"rejected_deferred=fail",
		"partial_no_write=pass",
		"partial_conflict=missing-source",
		"partial_removed_created=false",
		"partial_candidate_promoted=false",
		"deferred=generic gooo:invokes lifting",
	} {
		if !strings.Contains(evidence.Canonical(), line) {
			t.Fatalf("golden evidence omitted %q:\n%s", line, evidence.Canonical())
		}
	}
}
func TestBXEvidenceSortsLocalityIDsBeforeCanonicalization(t *testing.T) {
	evidence := BXEvidence{Locality: Locality{Touched: []ID{"b", "a"}, Affected: []ID{"d", "c"}}}
	if got := joinIDs(evidence.Locality.Touched); got != "a,b" {
		t.Fatalf("canonical locality is not sorted: %s", got)
	}
	if got := joinIDs(evidence.Locality.Affected); got != "c,d" {
		t.Fatalf("canonical closure is not sorted: %s", got)
	}
	if !reflect.DeepEqual(evidence.Locality.Touched, []ID{"b", "a"}) {
		t.Fatal("canonicalization mutated the evidence")
	}
}
func TestBXEvidenceRejectsMissingHardContract(t *testing.T) {
	fixture := incompleteBXFixture{}
	if _, err := MeasureBXFixture(fixture); err == nil || !strings.Contains(err.Error(), "hard BX evidence contract") {
		t.Fatalf("missing evidence contract was accepted: %v", err)
	}
	if _, err := MeasureBXFixture(missingArtifactsFixture{}); err == nil {
		t.Fatalf("missing base artifacts were accepted: %v", err)
	}
}
