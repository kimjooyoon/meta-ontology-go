package bodycodegen

import "testing"

func TestCollectionObservationUsesOrderedJSON7901(t *testing.T) {
	values := &recordBodySequence{Kind: 3, Values: []recordBodyScalar{
		{Kind: 3, Present: true, Integer: 9007199254741031},
		{Kind: 3, Present: true, Integer: -41},
	}}
	actual := recordBodyScalar{Kind: 6, Present: true, Collection: values}
	if got := recordScalarDisplay(actual); got != "[9007199254741031,-41]" {
		t.Fatal("collection observation must show source values", got)
	}
	empty := recordBodyScalar{Kind: 4, Present: true, Collection: &recordBodySequence{Kind: 1, Values: []recordBodyScalar{}}}
	if got := recordScalarDisplay(empty); got != "[]" {
		t.Fatal("empty collection observation", got)
	}
	wanted := actual
	wanted.Collection = &recordBodySequence{Kind: 3, Values: append([]recordBodyScalar(nil), values.Values...)}
	if !equalRecordBodyScalar(actual, wanted) {
		t.Fatal("equal independently stored arrays differ")
	}
	wanted.Collection.Values[0], wanted.Collection.Values[1] = wanted.Collection.Values[1], wanted.Collection.Values[0]
	if equalRecordBodyScalar(actual, wanted) {
		t.Fatal("reordered values compare as equal")
	}
	t.Log("7901: readable ordered JSON, exact int64, empty array, independent-storage equality and order-sensitive comparison; source execution0")
}
