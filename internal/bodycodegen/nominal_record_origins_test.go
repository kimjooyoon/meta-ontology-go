package bodycodegen

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

const nominalOriginSource = `package originpaths
namespace originpaths
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
entity Leaf id "origin://leaf" fields {
 field value id "origin://leaf/value" type integer required one
}
entity Pair id "origin://pair" fields {
 field before id "origin://pair/before" type Leaf required one
 field after id "origin://pair/after" type Leaf required one
}
activity Read(leaf Leaf) -> Integer id "origin://read" computes "return leaf.value"
activity After(pair Pair) -> Integer id "origin://after" computes "return pair.after.value"
activity Before(pair Pair) -> Integer id "origin://before" computes "return pair.before.value"
activity ThroughAlias(pair Pair) -> Integer id "origin://alias" computes "let moved = pair.after; return Read(moved)"
activity Both(pair Pair) -> Integer id "origin://both" computes "return pair.before.value + pair.after.value"
activity Copy(pair Pair) -> Integer id "origin://copy" computes "let moved=pair;return moved.after.value"
activity Replace(pair Pair, replacement Leaf) -> Integer id "origin://replace" computes "let moved=pair;moved.before=replacement;return moved.before.value"
activity Preserve(pair Pair, replacement Leaf) -> Integer id "origin://preserve" computes "let moved=pair;moved.before=replacement;return moved.after.value"
activity Choose(pair Pair, replacement Leaf, condition Boolean) -> Integer id "origin://choose" computes "let moved=pair;if condition {moved.before=replacement};return moved.before.value"
activity Rebuilt(pair Pair) -> Pair id "origin://rebuilt" computes "return Pair{before:pair.after,after:pair.before}"
activity Same(pair Pair) -> Boolean id "origin://same" computes "return pair.before==pair.after"
activity NewChild(pair Pair, replacement Integer) -> Integer id "origin://new-child" computes "let child=Leaf{value:replacement};let moved=pair;moved.before=child;return moved.before.value"
activity DefaultPair(pair Pair) -> Pair id "origin://default-pair" computes "return Pair{}"

`

func TestNominalRequiredSourceOriginPaths(t *testing.T) {
	for _, row := range []struct {
		activity string
		want     []string
	}{
		{"After", []string{"input:origin://pair/after > origin://leaf/value"}},
		{"Before", []string{"input:origin://pair/before > origin://leaf/value"}},
		{"ThroughAlias", []string{"input:origin://pair/after > origin://leaf/value"}},
		{"Both", []string{"input:origin://pair/after > origin://leaf/value", "input:origin://pair/before > origin://leaf/value"}},
	} {
		t.Run(row.activity, func(t *testing.T) {
			source := []byte(nominalOriginSource)
			r, err := InspectTypedValueOrigins("original.gooo", source, row.activity)
			if err != nil || r.Status != "RESOLVED" {
				t.Fatal("required nominal source path remains unresolved", err, r.Reason)
			}
			if r.SourceSHA256 != digest(source) || len(r.Outputs) != 1 {
				t.Fatal("lost source or return identity")
			}
			got := nominalOriginLeafPaths(r, r.Outputs[0].Root)
			if !reflect.DeepEqual(got, row.want) {
				t.Fatal("original container/leaf identities differ", got, row.want)
			}
			renamed := strings.ReplaceAll(nominalOriginSource, "field after id", "field later id")
			renamed = strings.ReplaceAll(renamed, "pair.after", "pair.later")
			changed, err := InspectTypedValueOrigins("relocated.gooo", []byte(renamed), row.activity)
			if err != nil || changed.Status != "RESOLVED" || !reflect.DeepEqual(nominalOriginLeafPaths(changed, changed.Outputs[0].Root), row.want) {
				t.Fatal("display rename or source filename changed origin paths", err, changed.Reason)
			}
		})
	}
}

func nominalOriginLeafPaths(r TypedValueOrigins, root uint16) []string {
	seen := make([]bool, len(r.Flow.Nodes)+1)
	seen[root] = true
	for _, n := range slices.Backward(r.Flow.Nodes) {
		if seen[n.ID] {
			seen[n.Parents[0]], seen[n.Parents[1]] = true, true
		}
	}
	var paths []string
	for _, in := range r.Inputs {
		if !seen[in.Node] || (in.FieldID != "origin://leaf/value" && in.FieldID != "") {
			continue
		}
		var parts []string
		id := in.Node
		for id != 0 {
			n := r.Flow.Nodes[id-1]
			if n.Kind != "input" {
				panic(fmt.Sprintf("input origin parent %d has kind %s", id, n.Kind))
			}
			if n.FieldID != "" {
				parts = append(parts, n.FieldID)
			}
			id = n.Parents[0]
		}
		slices.Reverse(parts)
		paths = append(paths, in.Port+":"+strings.Join(parts, " > "))
	}
	slices.Sort(paths)
	return paths
}

func TestNominalSourceCopyReplacementAndJoin(t *testing.T) {
	for _, row := range []struct {
		activity, field string
		want            []string
	}{
		{"Copy", "", []string{"input:origin://pair/after > origin://leaf/value"}},
		{"Replace", "", []string{"input1:origin://leaf/value"}},
		{"Preserve", "", []string{"input0:origin://pair/after > origin://leaf/value"}},
		{"Choose", "", []string{"input0:origin://pair/before > origin://leaf/value", "input1:origin://leaf/value"}},
		{"Rebuilt", "origin://pair/before", []string{"input:origin://pair/after > origin://leaf/value"}},
		{"Rebuilt", "origin://pair/after", []string{"input:origin://pair/before > origin://leaf/value"}},
		{"Same", "", []string{"input:origin://pair/after > origin://leaf/value", "input:origin://pair/before > origin://leaf/value"}},
		{"NewChild", "", []string{"input1:"}},
	} {
		t.Run(row.activity+row.field, func(t *testing.T) {
			r, err := InspectTypedValueOrigins("copies.gooo", []byte(nominalOriginSource), row.activity)
			if err != nil || r.Status != "RESOLVED" {
				t.Fatal(err, r.Reason)
			}
			var root uint16
			for _, out := range r.Outputs {
				if out.FieldID == row.field {
					root = out.Root
				}
			}
			if root == 0 {
				t.Fatal("missing output root")
			}
			if got := nominalOriginLeafPaths(r, root); !reflect.DeepEqual(got, row.want) {
				t.Fatal(got, row.want)
			}
		})
	}
}

func TestUnrepresentedNominalDefaultsCannotClaimComplete(t *testing.T) {
	r, err := InspectTypedValueOrigins("defaults.gooo", []byte(nominalOriginSource), "DefaultPair")
	if err != nil {
		t.Fatal(err)
	}
	if r.Status == "RESOLVED" {
		found := false
		for _, n := range r.Flow.Nodes {
			if n.FieldID == "origin://leaf/value" {
				found = true
			}
		}
		if !found {
			t.Fatal("complete nominal output omitted both default leaf field identities", r.Outputs)
		}
	} else if len(r.Outputs) != 0 || len(r.Flow.Nodes) != 0 {
		t.Fatal("unresolved defaults published an accepted prefix")
	}
}
