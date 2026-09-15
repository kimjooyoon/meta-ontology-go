package artifactcoverage

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/actionability"
)

func Evaluate(root string, action actionability.Report, observations ObservationDocument) (Report, error) {
	if err := validateRuntimeInput(action, observations); err != nil {
		return Report{}, err
	}
	program := CanonicalProgram()
	if err := Validate(program); err != nil {
		return Report{}, err
	}
	authorityDigest, err := AuthorityDigest(root, program)
	if err != nil {
		return Report{}, err
	}
	summary, witnesses, selected := evaluateOperations(action, observations, program.ArtifactBindings)
	indicators, err := evaluateIndicators(program.Indicators, summary)
	if err != nil {
		return Report{}, err
	}
	decision, reason := coverageDecision(action.Decision, summary)
	report := Report{
		Schema: ReportSchema, CommitSHA: observations.CommitSHA,
		Repository: observations.Repository, RunID: observations.RunID, RunAttempt: observations.RunAttempt,
		ActionabilityDigest: digestJSON(action), ObservationDigest: digestJSON(observations),
		ProgramDigest: digestJSON(program), AuthorityDigest: authorityDigest,
		Decision: decision, Reason: reason, SelectedOperation: selected,
		Summary: summary, Indicators: indicators, Operations: witnesses,
	}
	report.Proofs = coverageProofs(report)
	report.ReportDigest = digestJSON(report)
	return report, nil
}

func validateRuntimeInput(action actionability.Report, observations ObservationDocument) error {
	if action.Schema != actionability.Schema || len(action.CommitSHA) != 40 || action.Repository == "" {
		return fmt.Errorf("actionability identity is malformed")
	}
	if observations.Schema != ObservationSchema || observations.CommitSHA != action.CommitSHA ||
		observations.Repository != action.Repository || observations.RunID <= 0 ||
		observations.RunAttempt <= 0 || observations.RepositoryWrites < 0 {
		return fmt.Errorf("artifact observation identity is malformed")
	}
	return nil
}

func coverageDecision(actionDecision string, summary Summary) (string, string) {
	switch {
	case actionDecision != "FIXED_POINT":
		return "FAIL_CLOSED", "ACTIONABILITY_NOT_FIXED_POINT"
	case summary.RepositoryWrites != 0:
		return "FAIL_CLOSED", "ARTIFACT_OBSERVER_WRITE_EFFECT"
	case summary.AmbiguousOperations != 0:
		return "FAIL_CLOSED", "ARTIFACT_OBSERVATION_AMBIGUOUS"
	case summary.UncoveredOperations != 0:
		return "IMPROVE", "CANONICAL_ARTIFACT_COVERAGE_GAP"
	default:
		return "FIXED_POINT", "ALL_EXECUTABLE_OPERATIONS_ARTIFACT_BOUND"
	}
}
