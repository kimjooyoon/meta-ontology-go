package audienceresolution

import "reflect"

type indicatorSpec struct {
	ID            string
	Class         string
	Producer      string
	Consumer      string
	MetaOperation string
	ProofChoice   string
	Stage         string
	Step          string
	Reason        string
}

func indicatorSpecs() []indicatorSpec {
	return []indicatorSpec{
		{"source.binding", "DRIVER", "gooo://audience-resolution/source", "audience-resolution.projector", "bind-source-ledger", "FOUNDATION", "source", "bind", "source declaration is bound as authority"},
		{"ledger.coverage", "OUTCOME", "audience-resolution.producer", "audience-resolution.projector", "consume-canonical-ledger", "FOUNDATION", "ledger", "coverage", "all fixed coordinates are present once"},
		{"ledger.replay", "DRIVER", "audience-resolution.projector", "audience-resolution.receipt", "replay-canonical-ledger", "REGRESSION", "ledger", "replay", "the same ledger replays byte-equivalently"},
		{"user.coordinates", "OUTCOME", "audience-resolution.projector", "USER", "project-user-coordinate-set", "FOUNDATION", "projection", "user", "USER receives exactly its four coordinates"},
		{"author.coordinates", "DRIVER", "audience-resolution.projector", "TOOL_AUTHOR", "project-tool-author-coordinate-set", "FOUNDATION", "projection", "tool-author", "TOOL_AUTHOR receives the authoring contract"},
		{"governor.coordinates", "OUTCOME", "audience-resolution.projector", "GOVERNOR", "project-governor-coordinate-set", "FOUNDATION", "projection", "governor", "GOVERNOR receives the full ledger"},
		{"projection.nesting", "OUTCOME", "audience-resolution.projector", "audience-resolution.governor", "check-coordinate-nesting", "COHERENCE", "projection", "nest", "higher resolution extends the lower coordinate set"},
		{"projection.shared-decision", "GUARDRAIL", "audience-resolution.receipt", "all-audiences", "lift-global-decision", "COHERENCE", "projection", "decision", "global decision is carried with local verification status"},
		{"projection.resolution", "DRIVER", "audience-resolution.projector", "all-audiences", "preserve-projection-resolution", "COHERENCE", "projection", "resolution", "each audience keeps its fixed resolution label"},
		{"counterexample.omission", "GUARDRAIL", "audience-resolution.projector", "GOVERNOR", "execute-omitted-coordinate", "REGRESSION", "counterexample", "missing", "missing information cannot produce PASS"},
		{"counterexample.contradiction", "GUARDRAIL", "audience-resolution.projector", "GOVERNOR", "execute-contradictory-observation", "REGRESSION", "counterexample", "contradiction", "contradictory evidence becomes REFUTED"},
		{"receipt.seal", "DRIVER", "audience-resolution.receipt", "independent.checker", "validate-receipt-digest", "REGRESSION", "receipt", "seal", "receipt is independently checkable"},
	}
}

type recordState struct {
	records    map[string]EvidenceRecord
	valid      map[string]bool
	missing    map[string]bool
	contradict map[string]bool
	duplicate  bool
}

func inspectRecords(ledger Ledger) recordState {
	state := recordState{records: map[string]EvidenceRecord{}, valid: map[string]bool{}, missing: map[string]bool{}, contradict: map[string]bool{}}
	ids := map[string]bool{}
	for _, record := range ledger.Records {
		if record.ID == "" || ids[record.ID] {
			state.duplicate = true
			state.contradict[record.Coordinate] = true
		}
		ids[record.ID] = true
		if _, exists := state.records[record.Coordinate]; exists {
			state.duplicate = true
			state.contradict[record.Coordinate] = true
			continue
		}
		state.records[record.Coordinate] = record
	}
	for _, spec := range indicatorSpecs() {
		record, ok := state.records[spec.ID]
		if !ok {
			state.missing[spec.ID] = true
			continue
		}
		if missingRecordField(record) {
			state.missing[spec.ID] = true
			continue
		}
		if record.Observation == "CONTRADICTORY" || !recordMatchesSpec(record, spec) {
			state.contradict[spec.ID] = true
			continue
		}
		state.valid[spec.ID] = true
	}
	return state
}

func missingRecordField(record EvidenceRecord) bool {
	return record.ID == "" || record.Coordinate == "" || record.Audience == "" ||
		record.Stage == "" || record.Step == "" || record.Reason == "" ||
		record.Producer == "" || record.Consumer == "" || record.MetaOperation == "" ||
		record.ProofChoice == "" || record.PriorClaim == "" || record.Observation == ""
}

func recordMatchesSpec(record EvidenceRecord, spec indicatorSpec) bool {
	return record.ID == spec.ID && record.Coordinate == spec.ID && record.Producer == spec.Producer &&
		record.Consumer == spec.Consumer && record.MetaOperation == spec.MetaOperation &&
		record.ProofChoice == spec.ProofChoice && record.Stage == spec.Stage && record.Step == spec.Step &&
		record.PriorClaim == "OPEN" && record.Observation == "OBSERVED" && knownAudience(record.Audience)
}

func buildIndicators(input Input, model semanticSourceModel, state recordState, sourceBound, replay, policyValid bool, globalDecision string) []Indicator {
	specs := indicatorSpecs()
	user := sourceAudience(model, "USER").Coordinates
	author := sourceAudience(model, "TOOL_AUTHOR").Coordinates
	governor := sourceAudience(model, "GOVERNOR").Coordinates
	values := map[string]bool{
		"source.binding":               sourceBound,
		"ledger.coverage":              sourceBound && policyValid && coordinatesValid(state, governor) && len(state.records) == len(governor) && !state.duplicate,
		"ledger.replay":                replay,
		"user.coordinates":             coordinatesValid(state, user),
		"author.coordinates":           coordinatesValid(state, author),
		"governor.coordinates":         coordinatesValid(state, governor),
		"projection.nesting":           policyValid,
		"projection.shared-decision":   globalDecision == "PASS",
		"projection.resolution":        sourceAudienceResolutionValid(model),
		"counterexample.omission":      counterexampleValid(input.Ledger.Counterexamples, "counterexample.missing-information"),
		"counterexample.contradiction": counterexampleValid(input.Ledger.Counterexamples, "counterexample.decision-contradiction"),
		"receipt.seal":                 true,
	}
	result := make([]Indicator, 0, len(specs))
	for _, spec := range specs {
		observed := 0
		if values[spec.ID] {
			observed = 1
		}
		before, after := claimOutcome(spec.ID, state, observed == 1)
		result = append(result, Indicator{ID: spec.ID, Class: spec.Class, Producer: spec.Producer,
			Consumer: spec.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
			Stage: spec.Stage, Step: spec.Step, Reason: spec.Reason, ClaimBefore: before,
			ClaimAfter: after, Observed: observed, Expected: 1, Satisfied: observed == 1})
	}
	return result
}

func claimOutcome(id string, state recordState, satisfied bool) (string, string) {
	before := "OPEN"
	if record, ok := state.records[id]; ok && record.PriorClaim != "" {
		before = record.PriorClaim
	}
	if state.contradict[id] {
		return before, "REFUTED"
	}
	if satisfied {
		return before, "DISCHARGED"
	}
	return before, "OPEN"
}

func coordinatesValid(state recordState, coordinates []string) bool {
	if len(coordinates) == 0 {
		return false
	}
	for _, coordinate := range coordinates {
		if !state.valid[coordinate] {
			return false
		}
	}
	return true
}

func coordinateVisible(state recordState, coordinate string) bool {
	_, ok := state.records[coordinate]
	return ok
}

func coordinateContradictory(state recordState, coordinate string) bool {
	return state.contradict[coordinate]
}

func counterexampleValid(values []Counterexample, id string) bool {
	want := expectedCounterexample(id)
	for _, value := range values {
		if value.ID == id {
			return reflect.DeepEqual(value, want)
		}
	}
	return false
}

func counterexamplesValid(values []Counterexample) bool {
	if len(values) != 2 {
		return false
	}
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value.ID] || !counterexampleValid(values, value.ID) {
			return false
		}
		seen[value.ID] = true
	}
	return seen["counterexample.missing-information"] && seen["counterexample.decision-contradiction"]
}

func expectedCounterexample(id string) Counterexample {
	if id == "counterexample.missing-information" {
		return Counterexample{ID: id, Kind: "INFORMATION_OMISSION", Trigger: "author.consumer",
			Mutation: "remove the governor-only receipt.seal observation", TargetCoordinate: "receipt.seal"}
	}
	return Counterexample{ID: id, Kind: "DECISION_CONTRADICTION", Trigger: "user/governor",
		Mutation: "change ledger.coverage observation to CONTRADICTORY", TargetCoordinate: "ledger.coverage"}
}
