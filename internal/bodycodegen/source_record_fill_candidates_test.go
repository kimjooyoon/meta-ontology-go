package bodycodegen

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestRecordFillLiteralUsesDeclaredScalarType(t *testing.T) {
	for _, test := range []struct {
		name   string
		raw    string
		typeID string
		want   string
		ok     bool
	}{
		{name: "string", raw: `"ready"`, typeID: string(semantic.BuiltinStringTypeID), want: `"ready"`, ok: true},
		{name: "integer", raw: `42`, typeID: string(semantic.BuiltinIntegerTypeID), want: `42`, ok: true},
		{name: "boolean", raw: `true`, typeID: string(semantic.BuiltinBooleanTypeID), want: `true`, ok: true},
		{name: "decimal is not integer", raw: `4.2`, typeID: string(semantic.BuiltinIntegerTypeID)},
		{name: "wrong scalar type", raw: `42`, typeID: string(semantic.BuiltinStringTypeID)},
		{name: "null", raw: `null`, typeID: string(semantic.BuiltinBooleanTypeID)},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, ok := recordFillLiteral([]byte(test.raw), test.typeID)
			if got != test.want || ok != test.ok {
				t.Fatalf("record literal = %q, %v; want %q, %v", got, ok, test.want, test.ok)
			}
		})
	}
}
