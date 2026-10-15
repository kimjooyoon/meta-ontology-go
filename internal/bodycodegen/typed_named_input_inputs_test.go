package bodycodegen

import (
	"encoding/json"
	"reflect"
	"testing"
)

const namedColumnsSource = `package inputfacts
namespace inputfacts
entity Integer id "urn:gooo:type:integer"
entity Pair id "urn:inputfacts:pair" fields {
field left id "urn:inputfacts:left" type integer required one
field right id "urn:inputfacts:right" type integer required one
}
activity Delta(기준 Integer, Integer) -> Integer id "urn:inputfacts:delta" computes "return 기준 - input1"
activity Legacy(Integer) -> Integer id "urn:inputfacts:legacy" computes "return input + 29"
activity Read(쌍 Pair) -> Integer id "urn:inputfacts:read" computes "return 쌍.left + 쌍.right"
`

func namedColumnsReport(t *testing.T, activity string) TypedValueOrigins {
	t.Helper()
	r, err := InspectTypedValueOrigins("input-facts.gooo", []byte(namedColumnsSource), activity)
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	return r
}

func TestNamedColumnsSourceTransport(t *testing.T) {
	for activity, names := range map[string][]string{"Delta": {"기준", ""}, "Legacy": {""}, "Read": {"쌍"}} {
		r := namedColumnsReport(t, activity)
		raw, err := TypedNamedInputDeclarationInputs(r, []byte(namedColumnsSource), "source/control:DescribeInputs")
		if err != nil {
			t.Fatal(err)
		}
		var envelope struct {
			Inputs []map[string]NamedInputDeclarationColumns `json:"inputs"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil || len(envelope.Inputs) != 1 {
			t.Fatal(err)
		}
		c := envelope.Inputs[0]["source/control:DescribeInputs"]
		if c.Schema != "gooo/input-declaration-columns/v2" || c.Count != len(names) || !reflect.DeepEqual(c.Names, names) {
			t.Fatal(c)
		}
		for i, name := range names {
			if name == "" && (c.NameStarts[i] != 0 || c.NameEnds[i] != 0) ||
				name != "" && namedColumnsSource[c.NameStarts[i]:c.NameEnds[i]] != name {
				t.Fatal("name source range", c)
			}
		}
		assertLegacyColumnsStillPositional(t, r)
	}
}

func assertLegacyColumnsStillPositional(t *testing.T, r TypedValueOrigins) {
	t.Helper()
	raw, err := TypedInputDeclarationInputs(r, "source/control:InspectInputs")
	var envelope struct {
		Inputs []map[string]map[string]json.RawMessage
	}
	if err != nil || json.Unmarshal(raw, &envelope) != nil || len(envelope.Inputs) != 1 {
		t.Fatal("legacy column transport", err)
	}
	c := envelope.Inputs[0]["source/control:InspectInputs"]
	if string(c["schema"]) != `"gooo/input-declaration-columns/v1"` || len(c) != 11 {
		t.Fatal("legacy wire shape changed", c)
	}
}
