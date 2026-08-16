package semantic

import (
	"reflect"
	"testing"
)

func TestRelationRegistryIsClosedTypedAndDeterministic(t *testing.T) {
	want := []RelationSpec{
		{Predicate: Used, SubjectKind: Activity, ObjectKind: Entity},
		{Predicate: WasGeneratedBy, SubjectKind: Entity, ObjectKind: Activity},
		{Predicate: WasDerivedFrom, SubjectKind: Entity, ObjectKind: Entity},
		{Predicate: WasAssociatedWith, SubjectKind: Activity, ObjectKind: Agent},
	}
	first := RelationSpecs()
	second := RelationSpecs()
	if !reflect.DeepEqual(first, want) || !reflect.DeepEqual(second, want) {
		t.Fatalf("relation registry = %#v, want %#v", first, want)
	}
	first[0].SubjectKind = Agent
	if reflect.DeepEqual(first, RelationSpecs()) {
		t.Fatal("relation registry exposed mutable storage")
	}
	for _, spec := range want {
		got, ok := spec.Predicate.Spec()
		if !ok || got != spec {
			t.Fatalf("relation spec for %s = %#v, %t; want %#v", spec.Predicate, got, ok, spec)
		}
	}
	if Relation("wasAttributedTo").Valid() {
		t.Fatal("unsupported relation entered the closed registry")
	}
}
