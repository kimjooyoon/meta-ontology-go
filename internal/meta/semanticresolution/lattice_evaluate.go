package semanticresolution

const (
	LatticeSchema           = "gooo/meta-semantic-resolution-lattice/v1"
	LatticeCaseDenominator  = 4
	DecisionPass            = "PASS"
	DecisionFailClosed      = "FAIL_CLOSED"
	DecisionUnknown         = "UNKNOWN"
	DecisionLowerResolution = "LOWER_RESOLUTION"
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

func CanonicalLatticeCases() []LatticeCase {
	observations := []struct {
		id, claim   string
		observation PartialObservation
	}{
		{"exact-observation", "claim-exact-observation", PartialObservation{Required: 3, Observed: 3}},
		{"partial-invariant-descent", "claim-invariant-fallback", PartialObservation{Required: 3, Observed: 2, Reason: "REQUIRED_INPUT_UNOBSERVED"}},
		{"malformed-observation", "claim-exact-under-missing-evidence", PartialObservation{Required: 3, Observed: 4, Reason: "OBSERVATION_CARDINALITY_INVALID"}},
		{"mutation-authority", "claim-write-free-descent", PartialObservation{Required: 3, Observed: 2, Reason: "REQUIRED_INPUT_UNOBSERVED", MutationAuthority: true}},
	}
	cases := make([]LatticeCase, 0, len(observations))
	for _, item := range observations {
		transition := ResolvePartialObservation(item.observation)
		decision := transition.Decision
		if transition.Decision == DecisionLowerResolution {
			decision = DecisionUnknown
		}
		cases = append(cases, LatticeCase{ID: item.id, Decision: decision, Observation: item.observation, Transition: transition, ClaimID: item.claim})
	}
	return cases
}

func CanonicalClaims() []ClaimRecord {
	return []ClaimRecord{
		claim("claim-exact-observation", ClaimDischarged),
		claim("claim-invariant-fallback", ClaimOpen),
		claim("claim-exact-under-missing-evidence", ClaimRefuted),
		claim("claim-write-free-descent", ClaimDischarged),
	}
}

func claim(id string, state ClaimState) ClaimRecord {
	return ClaimRecord{ID: id, State: state, BeforeState: state, AfterState: state, Preserved: true}
}
