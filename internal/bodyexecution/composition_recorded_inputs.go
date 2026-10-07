package bodyexecution

import (
	"encoding/json"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func scalarInputTuple(input int64) []json.RawMessage {
	return []json.RawMessage{json.RawMessage(strconv.FormatInt(input, 10))}
}

func recordedInputTuples(report bodycodegen.Report) [][]json.RawMessage {
	var tuples [][]json.RawMessage
	if fill := report.BodyFill; fill != nil {
		tuples = append(tuples, integerCaseTuples(fill.SelectedCaseResults)...)
		tuples = append(tuples, integerCaseTuples(fill.HoldoutCaseResults)...)
		if probes := fill.BehavioralProbes; probes != nil {
			for _, input := range probes.ProbeInputs {
				tuples = append(tuples, scalarInputTuple(input))
			}
			for _, vector := range probes.ProbeVectors {
				tuples = append(tuples, integerInputTuple(vector))
			}
		}
	}
	if search := report.BodySearch; search != nil {
		tuples = append(tuples, integerCaseTuples(search.TrainingCaseResults)...)
		tuples = append(tuples, integerCaseTuples(search.HoldoutCaseResults)...)
	}
	if paths := report.BodyPaths; paths != nil {
		tuples = append(tuples, integerCaseTuples(paths.NativeCases)...)
		if observation := paths.Observation; observation != nil {
			for _, c := range observation.InitialCases {
				tuples = append(tuples, scalarInputTuple(c.Input))
			}
			for _, input := range observation.Options.Inputs {
				tuples = append(tuples, scalarInputTuple(input))
			}
			for _, round := range observation.Rounds {
				if round.Observation != nil {
					tuples = append(tuples, scalarInputTuple(round.Observation.Input))
				}
			}
		}
	}
	return tuples
}

func integerCaseTuples(cases []bodycodegen.IRBodyFillCaseResult) [][]json.RawMessage {
	tuples := make([][]json.RawMessage, 0, len(cases))
	for _, c := range cases {
		if len(c.Inputs) > 0 {
			tuples = append(tuples, integerInputTuple(c.Inputs))
		} else {
			tuples = append(tuples, scalarInputTuple(c.Input))
		}
	}
	return tuples
}

func integerInputTuple(inputs []int64) []json.RawMessage {
	tuple := make([]json.RawMessage, len(inputs))
	for i, input := range inputs {
		tuple[i] = json.RawMessage(strconv.FormatInt(input, 10))
	}
	return tuple
}
