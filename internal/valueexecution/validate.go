package valueexecution

import (
	"fmt"
	"regexp"
)

var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func Validate(report Report, headSHA string) error {
	if report.Schema != ReportSchema || report.HeadSHA != headSHA || !commitPattern.MatchString(headSHA) {
		return fmt.Errorf("value witness identity is invalid")
	}
	if report.Decision != DecisionProven || report.Reason != ReasonExactWitness || report.Resolution != ResolutionCoreValue {
		return fmt.Errorf("value witness failed closed: %s / %s / %s", report.Decision, report.Reason, report.Resolution)
	}
	if report.ValueProgram != "int.add:1" || report.Registry.RegisteredOperations != 1 || report.Registry.InvokedOperations != 1 {
		return fmt.Errorf("value program registry is not exact")
	}
	if !validDigest(report.SourceDigest) || !validDigest(report.ValueProgramDigest) || report.SemanticFingerprint == "" || report.CoreIRFingerprint == "" {
		return fmt.Errorf("value witness digests are invalid")
	}
	if err := requireExactCount("value cases", report.Summary.ValueCasesPassed, 5); err != nil {
		return err
	}
	if report.Summary.ValueCasesTotal != 5 || report.Summary.ValueOutputsObserved != 5 || report.Summary.DeterministicReplays != 5 {
		return fmt.Errorf("value case denominator or replay count changed")
	}
	if report.Summary.CounterexamplesPassed != 8 || report.Summary.CounterexamplesTotal != 8 {
		return fmt.Errorf("counterexample denominator changed")
	}
	if report.Improvement.ID != "value-level-computation" || report.Improvement.Before != coordinate(0, 1) || report.Improvement.After != coordinate(1, 1) || report.Improvement.BeforeEvidence != ReasonProgramMissing {
		return fmt.Errorf("improvement coordinate is not exact")
	}
	if report.Summary.CoreIRProgramPreserved != coordinate(1, 1) || report.Summary.CoreIRFingerprintSensitive != coordinate(1, 1) || report.Summary.CoreIRUnknownAttributeFailClosed != coordinate(1, 1) {
		return fmt.Errorf("core IR resolution boundary is not explicit")
	}
	if len(report.Indicators) != 18 || !allIndicatorsSatisfied(report.Indicators) || len(report.Views) != 3 || len(report.Proofs) != 3 {
		return fmt.Errorf("indicator or proof denominator changed")
	}
	expectedViews := []int{5, 14, 18}
	for index, view := range report.Views {
		if view.Total != expectedViews[index] || view.Satisfied != view.Total || view.BasisPoints != 10_000 {
			return fmt.Errorf("view %d is not exact", index)
		}
	}
	for _, proof := range report.Proofs {
		if !proof.Passed || !validDigest(proof.EvidenceDigest) {
			return fmt.Errorf("proof %s is invalid", proof.Choice)
		}
	}
	if len(report.NonClaims) != 5 || report.Summary.RepositoryWrites != 0 || report.Authority.RepositoryMutationAuthorized || report.Authority.PromotionAuthorized || report.Authority.AutomaticAdoptionAuthorized {
		return fmt.Errorf("non-claim or authority boundary changed")
	}
	if report.Digest != reportDigest(report) {
		return fmt.Errorf("value witness digest mismatch")
	}
	return nil
}
