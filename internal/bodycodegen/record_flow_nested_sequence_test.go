package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func nestedConstructedSequenceSource(t *testing.T) string {
	t.Helper()
	source := string(constructedSequenceSource(t, ""))
	source = strings.Replace(source, "entity Item", `entity Detail id "urn:construct:detail" fields {
 field amount id "urn:construct:amount" type integer required one
}
entity Item`, 1)
	source = strings.Replace(source, `field ignored id "urn:construct:ignored" type integer required one`,
		`field ignored id "urn:construct:ignored" type integer required one
 field details id "urn:construct:details" type Detail required many`, 1)
	source = strings.Replace(source, `field pick id "urn:construct:pick" type integer required one`,
		`field pick id "urn:construct:pick" type integer required one
 field sub id "urn:construct:sub" type integer required one`, 1)
	source = strings.Replace(source, "Item{value: input.left, ignored: input.noise}",
		"Item{value: input.noise, ignored: input.noise, details: []Detail{Detail{amount: input.left}}}", 1)
	return strings.Replace(source, "Item{value: input.right, ignored: input.noise}",
		"Item{value: input.noise, ignored: input.noise, details: []Detail{Detail{amount: input.right}}}", 1)
}

func nestedConstructedLoopSource(t *testing.T, emptyLeft, emptyRight bool) string {
	t.Helper()
	source := nestedConstructedSequenceSource(t)
	for _, row := range []struct {
		empty bool
		field string
	}{{emptyLeft, "left"}, {emptyRight, "right"}} {
		if row.empty {
			source = strings.Replace(source, "[]Detail{Detail{amount: input."+row.field+"}}", "[]Detail{}", 1)
		}
	}
	return strings.Replace(source, "return items[input.pick].value",
		"let total = 0\n for item in items { for detail in item.details { total = total + detail.amount } }\n return total", 1)
}

func TestConstructedSequenceNestedEmptyIteration(t *testing.T) {
	for _, row := range []struct {
		name                  string
		emptyLeft, emptyRight bool
		contributors          []string
		result                string
	}{
		{"nonempty", false, false, []string{"left", "right"}, "18"},
		{"empty-first", true, false, []string{"right"}, "11"},
		{"empty-second", false, true, []string{"left"}, "7"},
		{"both-empty", true, true, nil, "0"},
	} {
		t.Run(row.name, func(t *testing.T) {
			source := []byte(nestedConstructedLoopSource(t, row.emptyLeft, row.emptyRight))
			assertConstructedContributors(t, source, false, row.contributors...)
			assertConstructedContributors(t, source, true, row.contributors...)
			program, err := NewFixedProgram(context.Background(), "nested.gooo", source, "Read")
			if err != nil {
				t.Fatal(err)
			}
			got, err := program.Evaluate(context.Background(), []byte(`[{"left":7,"right":11,"noise":99,"pick":0,"sub":0}]`))
			if err != nil || strings.TrimSpace(string(got.Value)) != row.result {
				t.Fatal("empty iteration changed authored execution", got, err, row.result)
			}
		})
	}
}

func TestConstructedSequenceNestedIndexKeepsEmptyFault(t *testing.T) {
	source := nestedConstructedSequenceSource(t)
	source = strings.Replace(source, "items[input.pick].value", "items[input.pick].details[input.sub].amount", 1)
	assertConstructedContributors(t, []byte(source), false, "left", "right", "pick", "sub")
	for _, field := range []string{"left", "right"} {
		empty := strings.Replace(source, "[]Detail{Detail{amount: input."+field+"}}", "[]Detail{}", 1)
		r, err := InspectTypedValueOrigins("nested-index.gooo", []byte(empty), "Read")
		if err != nil || r.Status != "UNRESOLVED" || r.Reason != "FLOW_SEQUENCE_LITERAL_INDEX_BOUND" || len(r.Flow.Nodes) != 0 || len(r.Outputs) != 0 {
			t.Fatal("index fault became an accepted partial graph", field, err, r.Status, r.Reason)
		}
		selected, contributor := "1", "right"
		if field == "right" {
			selected, contributor = "0", "left"
		}
		valid := strings.Replace(empty, "items[input.pick].details[input.sub]", "items["+selected+"].details[0]", 1)
		assertConstructedContributors(t, []byte(valid), false, contributor)
	}
}

func TestConstructedSequenceNestedEmptyBranchKeepsControl(t *testing.T) {
	source := nestedConstructedLoopSource(t, true, false)
	source = strings.Replace(source, "let items = []Item{first, second}",
		"if input.pick != 0 { first.details = []Detail{Detail{amount: input.left}} }\n let items = []Item{first, second}", 1)
	assertConstructedContributors(t, []byte(source), false, "left", "right")
	assertConstructedContributors(t, []byte(source), true, "left", "right", "pick")
}
