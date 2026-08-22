package semanticresolution

const MaxResolutionDescents = 2

type Conflict struct {
	SourceDecision   string     `json:"source_decision"`
	CurrentResolution Resolution `json:"current_resolution"`
	Descents         int        `json:"descents"`
	RepositoryWrites int        `json:"repository_writes"`
}

type Transition struct {
	FromResolution   Resolution `json:"from_resolution"`
	ToResolution     Resolution `json:"to_resolution,omitempty"`
	Decision         string     `json:"decision"`
	Reason           string     `json:"reason"`
	NextOperation    string     `json:"next_operation,omitempty"`
	Descents         int        `json:"descents"`
	RepositoryWrites int        `json:"repository_writes"`
}

func CanonicalResolutions() []Resolution {
	return []Resolution{ResolutionExactOperation, ResolutionOperationClass, ResolutionInvariantOnly}
}

func LowerSemanticResolution(current Resolution) (Resolution, bool) {
	switch current {
	case ResolutionExactOperation:
		return ResolutionOperationClass, true
	case ResolutionOperationClass:
		return ResolutionInvariantOnly, true
	default:
		return "", false
	}
}

func ResolveSemanticConflict(conflict Conflict) Transition {
	result := Transition{FromResolution: conflict.CurrentResolution, Descents: conflict.Descents, RepositoryWrites: conflict.RepositoryWrites}
	switch {
	case conflict.RepositoryWrites != 0:
		result.Decision, result.Reason = "FAIL_CLOSED", "SEMANTIC_RESOLUTION_WRITE_EFFECT"
	case conflict.SourceDecision == "" || !validResolution(conflict.CurrentResolution) || conflict.Descents < 0:
		result.Decision, result.Reason = "FAIL_CLOSED", "SEMANTIC_CONFLICT_IDENTITY_INVALID"
	case conflict.Descents >= MaxResolutionDescents:
		result.Decision, result.Reason = "FAIL_CLOSED", "SEMANTIC_RESOLUTION_BUDGET_EXHAUSTED"
	default:
		next, ok := LowerSemanticResolution(conflict.CurrentResolution)
		if !ok {
			result.Decision, result.Reason = "FAIL_CLOSED", "SEMANTIC_RESOLUTION_EXHAUSTED"
			return result
		}
		result.ToResolution, result.Decision = next, "LOWER_RESOLUTION"
		result.Reason, result.NextOperation, result.Descents = "SEMANTIC_CONFLICT_COARSENED", "reevaluate-artifact-feedback", conflict.Descents+1
	}
	return result
}

func ReplayResolutionTransition(conflict Conflict) Transition {
	return ResolveSemanticConflict(conflict)
}

func validResolution(value Resolution) bool {
	return value == ResolutionExactOperation || value == ResolutionOperationClass || value == ResolutionInvariantOnly
}
