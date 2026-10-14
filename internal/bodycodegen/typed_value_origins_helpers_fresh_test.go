package bodycodegen

import (
	"reflect"
	"testing"
)

const typedSource13051 = `package helperorigins
namespace helperorigins
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
entity Pair id "urn:typed13051:pair" fields {
field base id "urn:typed13051:base" type integer required one
field next id "urn:typed13051:next" type integer required one
}
activity Carry(Pair) -> Pair id "urn:typed13051:carry" computes ` + "`let moved=input;moved.base=input.base+1493;return moved`" + `
activity Read(Pair) -> Integer id "urn:typed13051:read" computes ` + "`return input.next`" + `
activity Compose(Pair,Integer) -> Integer id "urn:typed13051:compose" computes ` + "`let threshold=1499;let staged=Carry(input0);staged.base=input1+threshold;return Read(staged)`" + `
activity Choose(Pair,Integer,Boolean) -> Pair id "urn:typed13051:choose" computes ` + "`let staged=Carry(input0);if input2 {staged.next=input1-1511}else{staged.base=input1+1523};return staged`" + `
activity Exit(Pair,Boolean) -> Integer id "urn:typed13051:exit" computes ` + "`if input1 {return input0.base};return input0.next`" + `
activity NoFlow(Integer) -> Integer id "urn:typed13051:no-flow" computes ` + "`let value=input;let i=0;for i=0;i<2;i=i+1 {value=value+1531};return value`" + `
`

func TestTypedHelperOrigins13051(t *testing.T) {
	for _, tc := range []struct {
		activity, field string
		control         bool
		want            []string
	}{
		{"Compose", "", false, []string{"input0:urn:typed13051:next"}},
		{"Choose", "urn:typed13051:next", false, []string{"input0:urn:typed13051:next", "input1:"}},
		{"Choose", "urn:typed13051:next", true, []string{"input0:urn:typed13051:next", "input1:", "input2:"}},
		{"Choose", "urn:typed13051:base", false, []string{"input0:urn:typed13051:base", "input1:"}},
		{"Exit", "", true, []string{"input0:urn:typed13051:base", "input0:urn:typed13051:next", "input1:"}},
	} {
		r, err := InspectTypedValueOrigins("helpers13051.gooo", []byte(typedSource13051), tc.activity)
		if err != nil || r.Status != "RESOLVED" {
			t.Fatal(tc.activity, err, r.Reason)
		}
		var root uint16
		for _, output := range r.Outputs {
			if output.FieldID == tc.field {
				root = output.Root
			}
		}
		if got := typedAncestors12973(r, root, tc.control); !reflect.DeepEqual(got, tc.want) {
			t.Fatal(tc.activity, tc.field, tc.control, got, tc.want)
		}
		if _, err := TypedValueOriginInputs(r, "helpers/reader:Inputs", tc.field, tc.control); err != nil {
			t.Fatal("topological export", tc.activity, err)
		}
	}
	r, err := InspectTypedValueOrigins("helpers13051.gooo", []byte(typedSource13051), "NoFlow")
	if err != nil || r.Status != "UNRESOLVED" || r.Reason != "FLOW_STATEMENT_UNSUPPORTED" || len(r.Outputs) != 0 || len(r.Flow.Nodes) != 0 {
		t.Fatal("bounded unresolved source", err, r)
	}
	t.Log("5 new field/helper/conditional-root assertions; literal local accepted before integer normalization; loop explicitly unresolved; native/model executions0")
}
