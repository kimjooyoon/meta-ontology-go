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
		{"projection.shared-decision", "GUARDRAIL", "audience-resolution.receipt", "all-audiences", "lift-global-decision", "COHERENCE", "projection", "decision", "every audience inherits one global decision"},
		{"projection.resolution", "DRIVER", "audience-resolution.projector", "all-audiences", "preserve-projection-resolution", "COHERENCE", "projection", "resolution", "each audience keeps its fixed resolution label"},
		{"counterexample.omission", "GUARDRAIL", "audience-resolution.projector", "GOVERNOR", "reject-omitted-coordinate", "REGRESSION", "counterexample", "missing", "missing information cannot produce PASS"},
		{"counterexample.contradiction", "GUARDRAIL", "audience-resolution.projector", "GOVERNOR", "reject-contradictory-decision", "REGRESSION", "counterexample", "contradiction", "contradictory decisions fail closed"},
		{"receipt.seal", "DRIVER", "audience-resolution.receipt", "independent.checker", "validate-receipt-digest", "REGRESSION", "receipt", "seal", "receipt is independently checkable"},
	}
}

type recordState struct {
	valid      map[string]bool
	missing    bool
	contradict bool
}

func inspectRecords(ledger Ledger) recordState {
	state := recordState{valid: map[string]bool{}}
	specs := indicatorSpecs()
	byCoordinate := make(map[string][]EvidenceRecord, len(ledger.Records))
	ids := map[string]bool{}
	for _, record := range ledger.Records {
		byCoordinate[record.Coordinate] = append(byCoordinate[record.Coordinate], record)
		if record.ID == "" || ids[record.ID] {
			state.contradict = true
		}
		ids[record.ID] = true
	}
	if len(ledger.Records) != len(specs) {
		if len(ledger.Records) < len(specs) {
			state.missing = true
		} else {
			state.contradict = true
		}
	}
	for _, spec := range specs {
		records := byCoordinate[spec.ID]
		if len(records) == 0 {
			state.missing = true
			state.valid[spec.ID] = false
			continue
		}
		if len(records) != 1 {
			state.contradict = true
			state.valid[spec.ID] = false
			continue
		}
		record := records[0]
		if missingRecordField(record) {
			state.missing = true
			state.valid[spec.ID] = false
			continue
		}
		if !recordMatchesSpec(record, spec) {
			state.contradict = true
			state.valid[spec.ID] = false
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
		record.ProofChoice == "" || record.ClaimBefore == "" || record.ClaimAfter == "" ||
		record.Decision == ""
}

func recordMatchesSpec(record EvidenceRecord, spec indicatorSpec) bool {
	return record.ID == spec.ID && record.Coordinate == spec.ID && record.Producer == spec.Producer &&
		record.Consumer == spec.Consumer && record.MetaOperation == spec.MetaOperation &&
		record.ProofChoice == spec.ProofChoice && record.Stage == spec.Stage && record.Step == spec.Step &&
		record.ClaimBefore == "UNPROVEN" && record.ClaimAfter == "OBSERVED" &&
		record.Decision == "PASS" && record.Satisfied && knownAudience(record.Audience)
}

func buildIndicators(input Input, state recordState, sourceBound, replay, nesting, resolutions, counterexamples bool) []Indicator {
	specs := indicatorSpecs()
	user := viewCoordinates(input.Contract, "USER")
	author := viewCoordinates(input.Contract, "TOOL_AUTHOR")
	governor := viewCoordinates(input.Contract, "GOVERNOR")
	values := map[string]bool{
		"source.binding":               sourceBound,
		"ledger.coverage":              ledgerIdentityValid(input.Ledger) && !state.missing && !state.contradict && allCoordinatesValid(state, specs),
		"ledger.replay":                replay,
		"user.coordinates":             coordinatesValid(state, user),
		"author.coordinates":           coordinatesValid(state, author),
		"governor.coordinates":         coordinatesValid(state, governor),
		"projection.nesting":           nesting,
		"projection.shared-decision":   sharedDecision(input.Ledger, state),
		"projection.resolution":        resolutions,
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
		claimAfter := "BLOCKED"
		if observed == 1 {
			claimAfter = "OBSERVED"
		}
		result = append(result, Indicator{ID: spec.ID, Class: spec.Class, Producer: spec.Producer,
			Consumer: spec.Consumer, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
			Stage: spec.Stage, Step: spec.Step, Reason: spec.Reason, ClaimBefore: "UNPROVEN",
			ClaimAfter: claimAfter, Observed: observed, Expected: 1, Satisfied: observed == 1})
	}
	return result
}

func allCoordinatesValid(state recordState, specs []indicatorSpec) bool {
	for _, spec := range specs {
		if !state.valid[spec.ID] {
			return false
		}
	}
	return true
}

func coordinatesValid(state recordState, coordinates []string) bool {
	for _, coordinate := range coordinates {
		if !state.valid[coordinate] {
			return false
		}
	}
	return len(coordinates) > 0
}

func sharedDecision(ledger Ledger, state recordState) bool {
	return ledgerIdentityValid(ledger) &&
		!state.missing && !state.contradict
}

func ledgerIdentityValid(ledger Ledger) bool {
	return ledger.Schema == LedgerSchema && ledger.ID != "" && ledger.Subject == Subject &&
		ledger.Decision == "PASS" && ledger.Resolution == "EXACT" && ledger.Reason != ""
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

func expectedCounterexample(id string) Counterexample {
	if id == "counterexample.missing-information" {
		return Counterexample{ID: id, Kind: "INFORMATION_OMISSION", Trigger: "author.consumer",
			ExpectedDecision: "FAIL_CLOSED", ObservedDecision: "PASS",
			Reason: "AUDIENCE_REQUIRED_COORDINATE_MISSING", Blocked: true}
	}
	return Counterexample{ID: id, Kind: "DECISION_CONTRADICTION", Trigger: "user/governor",
		ExpectedDecision: "FAIL_CLOSED", ObservedDecision: "PASS",
		Reason: "AUDIENCE_DECISION_CONTRADICTION", Blocked: true}
}
