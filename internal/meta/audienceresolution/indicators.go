package audienceresolution

const (
	EvidenceCurrent    = "CURRENT_EVIDENCE"
	EvidenceHistorical = "HISTORICAL_FIXTURE"
	EvidenceUnknown    = "UNKNOWN"
)

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
		{"ledger.coverage", "OUTCOME", "audience-resolution.producer", "audience-resolution.projector", "consume-canonical-ledger", "FOUNDATION", "ledger", "coverage", "all semantic coordinates have one raw recipe"},
		{"ledger.replay", "DRIVER", "audience-resolution.projector", "audience-resolution.receipt", "replay-canonical-ledger", "REGRESSION", "ledger", "replay", "two fresh producer executions replay byte-equivalently"},
		{"user.coordinates", "OUTCOME", "audience-resolution.projector", "USER", "project-user-coordinate-set", "FOUNDATION", "projection", "user", "USER receives exactly its four coordinates"},
		{"author.coordinates", "DRIVER", "audience-resolution.projector", "TOOL_AUTHOR", "project-tool-author-coordinate-set", "FOUNDATION", "projection", "tool-author", "TOOL_AUTHOR receives the authoring contract"},
		{"governor.coordinates", "OUTCOME", "audience-resolution.projector", "GOVERNOR", "project-governor-coordinate-set", "FOUNDATION", "projection", "governor", "GOVERNOR receives the full ledger"},
		{"projection.nesting", "OUTCOME", "audience-resolution.projector", "audience-resolution.governor", "check-coordinate-nesting", "COHERENCE", "projection", "nest", "higher resolution extends the lower coordinate set"},
		{"projection.shared-decision", "GUARDRAIL", "audience-resolution.receipt", "all-audiences", "lift-global-decision", "COHERENCE", "projection", "decision", "subject decision is carried with local verification status"},
		{"projection.resolution", "DRIVER", "audience-resolution.projector", "all-audiences", "preserve-projection-resolution", "COHERENCE", "projection", "resolution", "each audience keeps its fixed resolution label"},
		{"counterexample.omission", "GUARDRAIL", "audience-resolution.projector", "GOVERNOR", "execute-omitted-coordinate", "REGRESSION", "counterexample", "missing", "executed omission changes the target evidence and decision"},
		{"counterexample.contradiction", "GUARDRAIL", "audience-resolution.projector", "GOVERNOR", "execute-contradictory-observation", "REGRESSION", "counterexample", "executed contradiction changes the target evidence and decision"},
		{"receipt.seal", "DRIVER", "audience-resolution.receipt", "independent.checker", "validate-receipt-digest", "REGRESSION", "receipt", "seal", "receipt is independently checkable after subject evaluation"},
	}
}

func subjectIndicatorIDs() map[string]bool {
	return map[string]bool{
		"source.binding": true, "ledger.coverage": true, "ledger.replay": true,
		"user.coordinates": true, "author.coordinates": true, "governor.coordinates": true,
		"projection.nesting": true, "projection.resolution": true,
		"counterexample.omission": true, "counterexample.contradiction": true,
	}
}

func subjectCoordinates(model semanticSourceModel) map[string]bool {
	result := map[string]bool{}
	for _, coordinate := range sourceCoordinates(model) {
		if coordinate != "projection.shared-decision" && coordinate != "receipt.seal" {
			result[coordinate] = true
		}
	}
	return result
}

type recordState struct {
	records    map[string]EvidenceRecord
	valid      map[string]bool
	missing    map[string]bool
	contradict map[string]bool
	visible    map[string]bool
	duplicate  bool
}

func inspectCurrentEvidence(recipes []EvidenceRecord, current []EvidenceRecord) recordState {
	state := recordState{records: map[string]EvidenceRecord{}, valid: map[string]bool{}, missing: map[string]bool{}, contradict: map[string]bool{}, visible: map[string]bool{}}
	recipeMap := recordMap(recipes)
	ids := map[string]bool{}
	for _, record := range current {
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
	for _, recipe := range recipes {
		if record, ok := state.records[recipe.Coordinate]; ok {
			state.visible[recipe.Coordinate] = true
			if record.EvidenceStatus == EvidenceCurrent && record.ObservedValue == "false" {
				state.contradict[recipe.Coordinate] = true
				continue
			}
			if record.EvidenceStatus != EvidenceCurrent || record.ObservedValue != "true" {
				state.missing[recipe.Coordinate] = true
				continue
			}
			if !currentMatchesRecipe(record, recipe) {
				state.contradict[recipe.Coordinate] = true
				continue
			}
			state.valid[recipe.Coordinate] = true
			continue
		}
		state.missing[recipe.Coordinate] = true
	}
	for coordinate := range recipeMap {
		if _, ok := state.records[coordinate]; !ok {
			state.visible[coordinate] = false
		}
	}
	return state
}

func currentMatchesRecipe(current, recipe EvidenceRecord) bool {
	return current.ID == recipe.ID && current.Coordinate == recipe.Coordinate && current.ClaimID == recipe.ClaimID &&
		current.Proposition == recipe.Proposition && current.PropositionDigest == recipe.PropositionDigest &&
		current.TargetAddress == recipe.TargetAddress && current.Provider != "" && current.ArtifactPath != "" &&
		validDigest(current.ContentDigest) && current.ObservedPredicate != "" && current.EvidenceStatus == EvidenceCurrent &&
		current.Producer == recipe.Producer && current.Consumer == recipe.Consumer && current.MetaOperation == recipe.MetaOperation &&
		current.ProofChoice == recipe.ProofChoice && current.Stage == recipe.Stage && current.Step == recipe.Step && current.Reason == recipe.Reason &&
		current.PriorClaim == recipe.PriorClaim
}

func recordMap(records []EvidenceRecord) map[string]EvidenceRecord {
	result := map[string]EvidenceRecord{}
	for _, record := range records {
		if _, exists := result[record.Coordinate]; !exists {
			result[record.Coordinate] = record
		}
	}
	return result
}

func buildIndicators(recipes []EvidenceRecord, current []EvidenceRecord, state recordState, model semanticSourceModel, replay ReplayVerification, cex []CounterexampleResult, sourceBound, policyValid bool, subjectDecision string) []Indicator {
	currentMap := recordMap(current)
	values := map[string]bool{
		"source.binding":               state.valid["source.binding"],
		"ledger.coverage":              state.valid["ledger.coverage"],
		"ledger.replay":                state.valid["ledger.replay"] && replay.Equal,
		"user.coordinates":             coordinatesValid(state, sourceAudience(model, "USER").Coordinates),
		"author.coordinates":           coordinatesValid(state, sourceAudience(model, "TOOL_AUTHOR").Coordinates),
		"governor.coordinates":         coordinatesValid(state, sourceAudience(model, "GOVERNOR").Coordinates),
		"projection.nesting":           policyValid,
		"projection.shared-decision":   subjectDecision == "PASS",
		"projection.resolution":        sourceAudienceResolutionValid(model),
		"counterexample.omission":      counterexamplePassed(cex, "counterexample.missing-information"),
		"counterexample.contradiction": counterexamplePassed(cex, "counterexample.decision-contradiction"),
		// receipt.seal is intentionally excluded from subject PASS. It is
		// discharged only by the post-evaluation independent attestation.
		"receipt.seal": false,
	}
	result := make([]Indicator, 0, len(indicatorSpecs()))
	for _, spec := range indicatorSpecs() {
		recipe := recipeFor(recipes, spec.ID)
		currentRecord := currentMap[spec.ID]
		observed := boolToInt(values[spec.ID])
		before, after := claimOutcome(spec.ID, state, observed == 1)
		status := EvidenceUnknown
		if currentRecord.EvidenceStatus != "" {
			status = currentRecord.EvidenceStatus
		}
		result = append(result, Indicator{ID: spec.ID, Class: spec.Class, Producer: recipe.Producer,
			Consumer: recipe.Consumer, MetaOperation: recipe.MetaOperation, ProofChoice: recipe.ProofChoice,
			Stage: recipe.Stage, Step: recipe.Step, Reason: recipe.Reason, ClaimBefore: before,
			ClaimAfter: after, Observed: observed, Expected: 1, Satisfied: observed == 1,
			EvidenceStatus: status, PropositionDigest: recipe.PropositionDigest, TargetAddress: recipe.TargetAddress,
			ArtifactPath: currentRecord.ArtifactPath, ContentDigest: currentRecord.ContentDigest})
	}
	_ = sourceBound
	return result
}

func recipeFor(recipes []EvidenceRecord, coordinate string) EvidenceRecord {
	for _, recipe := range recipes {
		if recipe.Coordinate == coordinate {
			return recipe
		}
	}
	proposition := "semantic policy coordinate " + coordinate
	return EvidenceRecord{ID: coordinate, Coordinate: coordinate, ClaimID: "claim/" + coordinate,
		Proposition: proposition, PropositionDigest: digestBytes([]byte(proposition)), TargetAddress: "gooo://audience-resolution/claim/" + coordinate,
		Provider: "audience-resolution.policy", Producer: "audience-resolution.policy", Consumer: "audience-resolution.policy",
		MetaOperation: "project-audience-claim", ProofChoice: "COHERENCE", Stage: "projection", Step: "policy",
		Reason: "formal source policy coordinate has no raw recipe", PriorClaim: "OPEN"}
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

func coordinateVisible(state recordState, coordinate string) bool { return state.visible[coordinate] }

func coordinateContradictory(state recordState, coordinate string) bool {
	return state.contradict[coordinate]
}

func counterexamplePassed(values []CounterexampleResult, id string) bool {
	for _, value := range values {
		if value.ID == id {
			return value.ExecutionValidated
		}
	}
	return false
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
