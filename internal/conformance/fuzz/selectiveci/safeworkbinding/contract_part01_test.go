package safeworkbinding

import (
	"encoding/hex"
	"reflect"
	"testing"
)

type fieldSpec struct {
	name string
	typ  reflect.Type
	tag  string
}

func checkFields(t *testing.T, typ reflect.Type, want []fieldSpec) {
	t.Helper()
	if typ.NumField() != len(want) {
		t.Fatalf("field count: got %d want %d", typ.NumField(), len(want))
	}
	for i, expected := range want {
		field := typ.Field(i)
		if field.Name != expected.name || field.Type != expected.typ {
			t.Fatalf("field %d: got %s %s", i, field.Name, field.Type)
		}
		if field.Tag.Get("json") != expected.tag {
			t.Fatalf("field %s: json tag %q", field.Name, field.Tag.Get("json"))
		}
	}
}
func check(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Fatal(message)
	}
}
func checkField(t *testing.T, got frameField, tag frameTag, value []byte) {
	t.Helper()
	check(t, got.tag == tag, "field tag")
	check(t, hex.EncodeToString(got.value) == hex.EncodeToString(value), "field value")
}
