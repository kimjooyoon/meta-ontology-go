package selfimprovementobservation

import (
	"bytes"
	"encoding/json"
)

func Build(in Inputs, opts Options) Observation {
	left, right := project(in, opts), project(in, opts)
	leftBytes, _ := json.Marshal(left.Observation)
	rightBytes, _ := json.Marshal(right.Observation)
	left.Observation.Indicators = observationIndicators(left.Validation, bytes.Equal(leftBytes, rightBytes))
	left.Observation.Views = observationViews(left.Observation.Indicators)
	left.Observation.Proofs = observationProofs(left.Observation.Indicators, left.Observation.InputDigest)
	finishObservation(&left.Observation, left.Validation, in.Report.Value)
	canonical := left.Observation
	canonical.Digest = ""
	left.Observation.Digest = digestJSON(canonical)
	return left.Observation
}

func finishObservation(observation *Observation, check validation, source SourceReport) {
	satisfied := 0
	for _, indicator := range observation.Indicators {
		if indicator.Satisfied {
			satisfied++
		}
	}
	observation.Summary.Coordinates = CountSummary{Satisfied: satisfied, Total: len(observation.Indicators), BasisPoints: basisPoints(satisfied, len(observation.Indicators))}
	if satisfied == len(observation.Indicators) {
		observation.Decision, observation.Resolution = "OBSERVED", "EXACT"
		observation.Reason, observation.Interpretation = "READ_ONLY_MINIMAL_VALUE_BOUND", "READ_ONLY_IMPROVEMENT_INPUT"
		return
	}
	observation.Decision, observation.Interpretation = "FAIL_CLOSED", "NO_IMPROVEMENT_INPUT"
	observation.Resolution, observation.Reason = failureClassification(check, source)
	if observation.Resolution == "LOWER_RESOLUTION" {
		observation.Summary.Unknowns = 1
	}
}

func basisPoints(satisfied, total int) int {
	if total == 0 {
		return 0
	}
	return satisfied * 10000 / total
}
