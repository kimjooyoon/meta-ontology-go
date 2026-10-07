package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type compositionSelectionInputs [compositionLimit]map[string]struct{}

func (graph compositionGraph) selectionInputs(ctx context.Context, filename string, source []byte,
	prior Composition) (compositionSelectionInputs, error) {
	var sets compositionSelectionInputs
	if len(prior.Steps) != graph.count {
		return sets, fmt.Errorf("composition selection step count differs")
	}
	for i, node := range graph.nodes[:graph.count] {
		if !node.Assembling {
			continue
		}
		if search := prior.Steps[i].Generation.Report.BodySearch; search != nil && search.ExternalTrainingFeedback != nil {
			return sets, fmt.Errorf("external selection input tuples are unavailable")
		}
		spec, err := bodycodegen.SourceAssembly(ctx, filename, source, node.Name)
		if err != nil || spec == nil {
			return sets, fmt.Errorf("source assembly inputs unavailable for %s", node.Name)
		}
		var tuples [][]json.RawMessage
		for _, cases := range [][]assemblyspec.Case{spec.Cases, spec.HoldoutCases} {
			for _, c := range cases {
				tuples = append(tuples, scalarInputTuple(c.Input))
			}
		}
		for _, cases := range [][]assemblyspec.ValueCase{spec.ValueCases, spec.ValueHoldoutCases} {
			for _, c := range cases {
				var tuple []json.RawMessage
				if err := json.Unmarshal([]byte(c.Inputs), &tuple); err != nil {
					return sets, err
				}
				tuples = append(tuples, tuple)
			}
		}
		tuples = append(tuples, recordedInputTuples(prior.Steps[i].Generation.Report)...)
		sets[i] = make(map[string]struct{}, len(tuples))
		for _, tuple := range tuples {
			key, err := graph.selectionInputKey(node, tuple)
			if err != nil {
				return sets, err
			}
			sets[i][key] = struct{}{}
		}
	}
	return sets, nil
}

func (graph compositionGraph) selectionInputKey(node CompositionActivity, tuple []json.RawMessage) (string, error) {
	slots := node.inputSlots()
	if len(tuple) != len(slots) {
		return "", fmt.Errorf("selection input arity differs")
	}
	values := make([]json.RawMessage, len(tuple))
	for i, input := range tuple {
		canonical, err := graph.canonicalValue(input, slots[i].Type)
		if err != nil {
			return "", err
		}
		values[i] = canonical
	}
	raw, err := json.Marshal(values)
	return string(raw), err
}
