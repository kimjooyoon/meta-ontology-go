package bodycodegen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

const typedSource12973 = `package typedfresh
namespace typedfresh
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
entity Pair id "urn:typed12973:pair" fields {
field left id "urn:typed12973:left" type integer required one
field right id "urn:typed12973:right" type integer required one
}
activity Read(Pair) -> Integer id "urn:typed12973:read" computes "return input.\u0072ight"
activity Copied(Pair,Integer) -> Integer id "urn:typed12973:copied" computes ` + "`let original=input0;let copied=original;copied.left=input1+1451;return Read(copied)`" + `
activity Original(Pair,Integer) -> Integer id "urn:typed12973:original" computes ` + "`let original=input0;let copied=original;copied.left=input1+1453;return original.left`" + `
activity Updated(Pair,Integer) -> Pair id "urn:typed12973:updated" computes ` + "`let copied=input0;copied.left=input1+1459;return copied`" + `
activity Choice(Pair,Integer,Boolean) -> Integer id "urn:typed12973:choice" computes ` + "`let copied=input0;if input2 {copied.right=input1+1471};return Read(copied)`" + `
activity Constant(Pair,Integer) -> Integer id "urn:typed12973:constant" computes ` + "`if false {return input1+1481};return Read(input0)`" + `
activity Loop(Integer) -> Integer id "urn:typed12973:loop" computes ` + "`let result=input;let i=0;for i=0;i<3;i=i+1 {result=result+1483};return result`" + `
`

func TestFreshTypedOrigins12973(t *testing.T) {
	source := []byte(typedSource12973)
	checks := []struct {
		activity, field string
		control         bool
		want            []string
	}{
		{"Copied", "", false, []string{"input0:urn:typed12973:right"}},
		{"Original", "", false, []string{"input0:urn:typed12973:left"}},
		{"Updated", "urn:typed12973:left", false, []string{"input1:"}},
		{"Updated", "urn:typed12973:right", false, []string{"input0:urn:typed12973:right"}},
		{"Choice", "", false, []string{"input0:urn:typed12973:right", "input1:"}},
		{"Choice", "", true, []string{"input0:urn:typed12973:right", "input1:", "input2:"}},
		{"Constant", "", true, []string{"input0:urn:typed12973:right"}},
	}
	for _, check := range checks {
		t.Run(check.activity+check.field+map[bool]string{true: "control", false: "data"}[check.control], func(t *testing.T) {
			r, err := InspectTypedValueOrigins("escaped12973.gooo", source, check.activity)
			if err != nil || r.Status != "RESOLVED" {
				t.Fatal(err, r.Reason)
			}
			var root uint16
			for _, output := range r.Outputs {
				if output.FieldID == check.field {
					root = output.Root
				}
			}
			got := typedAncestors12973(r, root, check.control)
			if !reflect.DeepEqual(got, check.want) {
				t.Fatal("input ancestry", got, check.want)
			}
			for i, node := range r.Flow.Nodes {
				if node.Span.ActivityID == "" || node.Span.View == "input-port" {
					continue
				}
				location := r.Sources[i]
				if location.Status != "MAPPED" || location.End > len(source) || location.SourceSHA256 != digest(source) {
					t.Fatal("original source range", node, location)
				}
				if node.Span.ActivityID == "urn:typed12973:read" && !strings.Contains(string(source[location.Start:location.End]), `\u0072ight`) {
					t.Fatal("escaped helper source", location)
				}
			}
			if _, err := TypedValueOriginInputs(r, "typed/consumer:Inputs", check.field, check.control); err != nil {
				t.Fatal(err)
			}
			if dir := os.Getenv("GOOO_TYPED_OUTPUT"); dir != "" {
				raw, _ := json.MarshalIndent(r, "", "  ")
				if err := os.WriteFile(filepath.Join(dir, strings.ReplaceAll(check.activity+check.field+map[bool]string{true: "control", false: "data"}[check.control], ":", "_")+".json"), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	r, err := InspectTypedValueOrigins("loop12973.gooo", source, "Loop")
	if err != nil || r.Status != "UNRESOLVED" || r.Reason != "FLOW_STATEMENT_UNSUPPORTED" || len(r.Outputs) != 0 || len(r.Flow.Nodes) != 0 {
		t.Fatal("unsupported loop must not publish a resolved prefix", err, r)
	}
	if _, err := TypedValueOriginInputs(r, "typed/consumer:Inputs", "", false); err == nil {
		t.Fatal("unresolved graph exported")
	}
	t.Log("7 source/field/control ancestry assertions; escaped helper bytes checked; loop unresolved without partial graph; no cases, model or native execution")
}

func typedAncestors12973(r TypedValueOrigins, root uint16, control bool) []string {
	seen := make([]bool, len(r.Flow.Nodes)+1)
	seen[root] = true
	for _, n := range slices.Backward(r.Flow.Nodes) {
		if !seen[n.ID] {
			continue
		}
		seen[n.Parents[0]], seen[n.Parents[1]] = true, true
		if control {
			seen[n.Guard], seen[n.Condition] = true, true
		}
	}
	var result []string
	for _, in := range r.Inputs {
		if seen[in.Node] {
			result = append(result, in.Port+":"+in.FieldID)
		}
	}
	slices.Sort(result)
	return result
}
