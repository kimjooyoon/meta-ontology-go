package bodycodegen

import (
	"fmt"
	"go/token"
	"sort"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func (b *recipeArena) choice(spec sourceRecipeChoice, root []int) (pathplan.Choice, error) {
	choice := pathplan.Choice{ID: spec.ID, Kind: spec.Kind, Intent: spec.Intent}
	if spec.Occurrence == nil || *spec.Occurrence < 0 || *spec.Occurrence >= 128 {
		return choice, fmt.Errorf("recipe choice requires an explicit occurrence in 0..127")
	}
	if spec.Kind == pathplan.RootOrder {
		if spec.Alternative != "" || *spec.Occurrence+1 >= len(root) {
			return choice, fmt.Errorf("root order requires two adjacent root statements")
		}
		reverse := append([]int(nil), root...)
		i := *spec.Occurrence
		reverse[i], reverse[i+1] = reverse[i+1], reverse[i]
		choice.Fallback = "schedule_forward"
		choice.Options = []pathplan.Option{{Label: choice.Fallback, Order: append([]int(nil), root...)},
			{Label: "schedule_reverse", Order: reverse}}
		return choice, nil
	}
	target, err := b.target(spec.Kind, *spec.Occurrence)
	if err != nil {
		return choice, err
	}
	choice.Target = target
	switch spec.Kind {
	case pathplan.OperandOrder, pathplan.BranchLayout:
		if spec.Alternative != "" {
			return choice, fmt.Errorf("layout recipe must not name an alternative local")
		}
		choice.Fallback = "layout_forward"
		choice.Options = []pathplan.Option{{Label: choice.Fallback}, {Label: "layout_reverse", Reverse: true}}
	case pathplan.LocalReference, pathplan.AssignmentTarget:
		name, first, second := b.expressions[target].Name, "reference_first", "reference_second"
		if spec.Kind == pathplan.AssignmentTarget {
			name, first, second = b.statements[target].Name, "assign_first", "assign_second"
		}
		if spec.Alternative == "" || spec.Alternative == name {
			return choice, fmt.Errorf("name recipe requires a distinct alternative local")
		}
		choice.Fallback = first
		choice.Options = []pathplan.Option{{Label: first, Name: name}, {Label: second, Name: spec.Alternative}}
	}
	return choice, nil
}

// Source-order selectors are independent of arena allocation order. When nested
// nodes begin at the same position, the containing node precedes its child.
func (b *recipeArena) target(kind string, occurrence int) (int, error) {
	var positions [128]struct {
		index int
		pos   token.Pos
	}
	count := 0
	for i := range 128 {
		var pos token.Pos
		switch kind {
		case pathplan.OperandOrder:
			if i < b.expressionCount && b.expressions[i].Kind == bodyplan.ExprBinary {
				pos = b.expressionPositions[i]
			}
		case pathplan.LocalReference:
			if i < b.expressionCount && b.expressions[i].Kind == bodyplan.ExprLocal {
				pos = b.expressionPositions[i]
			}
		case pathplan.AssignmentTarget, pathplan.BranchLayout:
			want := bodyplan.StmtAssign
			if kind == pathplan.BranchLayout {
				want = bodyplan.StmtIf
			}
			if i < b.statementCount && b.statements[i].Kind == want {
				pos = b.statementPositions[i]
			}
		default:
			return 0, fmt.Errorf("unknown recipe choice kind %q", kind)
		}
		if pos != token.NoPos {
			positions[count].index, positions[count].pos = i, pos
			count++
		}
	}
	if occurrence >= count {
		return 0, fmt.Errorf("recipe %s occurrence %d exceeds %d source sites", kind, occurrence, count)
	}
	sort.Slice(positions[:count], func(i, j int) bool {
		if positions[i].pos == positions[j].pos {
			return positions[i].index > positions[j].index
		}
		return positions[i].pos < positions[j].pos
	})
	return positions[occurrence].index, nil
}
