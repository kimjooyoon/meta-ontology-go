package semanticresolution

const (
	LatticeSchema                    = "gooo/meta-semantic-resolution-lattice/v1"
	LatticeCaseDenominator           = 4
	LatticeCounterfactualDenominator = 2
	DecisionPass                     = "PASS"
	DecisionFailClosed               = "FAIL_CLOSED"
	DecisionUnknown                  = "UNKNOWN"
	DecisionLowerResolution          = "LOWER_RESOLUTION"
)

func ResolvePartialObservation(observation PartialObservation) LatticeTransition {
	result := LatticeTransition{
		FromResolution:    ResolutionExactOperation,
		RepositoryWrites:  observation.RepositoryWrites,
		MutationAuthority: observation.MutationAuthority,
	}
	switch {
	case observation.RepositoryWrites != 0:
		result.Decision, result.Reason = DecisionFailClosed, "REPOSITORY_WRITE_EFFECT"
	case observation.MutationAuthority:
		result.Decision, result.Reason = DecisionFailClosed, "MUTATION_AUTHORITY_PRESENT"
	case observation.Required <= 0 || observation.Observed < 0 || observation.Observed > observation.Required:
		result.Decision, result.Reason = DecisionFailClosed, "OBSERVATION_CARDINALITY_INVALID"
	case observation.Observed == observation.Required:
		result.ToResolution, result.Decision, result.Reason = ResolutionExactOperation, DecisionPass, "OBSERVATION_COMPLETE"
	default:
		result.ToResolution, result.Decision, result.Reason = ResolutionInvariantOnly, DecisionLowerResolution, "PARTIAL_OBSERVATION"
		result.Unknown = &UnknownValue{Stage: StagePartialObservation, Step: 1, Reason: observation.Reason}
	}
	return result
}

func ReplayPartialObservation(observation PartialObservation) LatticeTransition {
	return ResolvePartialObservation(observation)
}
