package bidir

import (
	"reflect"
	"testing"
)

func TestPutRejectsInvalidSourceWithoutWrite(t *testing.T) {
	document := billingDocument()
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	invalid := document
	invalid.Relations = append(invalid.Relations, Relation{
		Kind:   PredicateWasDerivedFrom,
		Source: "billing://entity/missing",
		Target: "billing://entity/order",
	})

	written, err := Put(invalid, model)
	if err == nil {
		t.Fatal("Put accepted an invalid source document")
	}
	if !reflect.DeepEqual(written, invalid) {
		t.Fatalf("invalid source was not returned unchanged: %#v", written)
	}
	assertPutNoWrite(t, err, PutSourceInvalid)
}
func TestPutRejectsMissingSpanWithoutWrite(t *testing.T) {
	document := billingDocument()
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	updated := base.Clone()
	updated.Relations = append(updated.Relations, Relation{
		Kind:   PredicateWasDerivedFrom,
		Source: "billing://entity/payment",
		Target: "billing://entity/order",
	})

	written, err := Put(document, updated)
	if err == nil {
		t.Fatal("Put accepted a semantic update without provenance")
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatalf("missing-span update changed the source view: %#v", written)
	}
	assertPutNoWrite(t, err, PutProvenanceMissing)
}
func TestPutRejectsConflictingRelationWithoutWrite(t *testing.T) {
	document := billingDocument()
	conflicting := document
	conflicting.Relations = []Relation{
		{Kind: PredicateWasDerivedFrom, Source: "billing://entity/payment", Target: "billing://entity/order"},
		{Kind: PredicateWasDerivedFrom, Source: "billing://entity/payment", Target: "billing://entity/order", Attributes: map[string]string{"source": "conflict"}},
	}
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}

	written, err := Put(conflicting, base)
	if err == nil {
		t.Fatal("Put accepted conflicting source relations")
	}
	if !reflect.DeepEqual(written, conflicting) {
		t.Fatalf("conflicting source was not returned unchanged: %#v", written)
	}
	assertPutNoWrite(t, err, PutSourceInvalid)
}
