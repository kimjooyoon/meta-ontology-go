package valuecatalog

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func ValidateUnknown(report Report, headSHA string) error {
	if report.Schema != ReportSchema || report.HeadSHA != headSHA || !commitPattern.MatchString(headSHA) {
		return fmt.Errorf("unknown catalog identity is invalid")
	}
	wantCoordinate := ProcessCoordinate{Stage: "RESOLVE", Step: "resolve-operation-spec", Reason: valueexecution.ReasonProgramUnknown}
	if report.Decision != DecisionFailClosed || report.Reason != ReasonObservationFailed ||
		report.Resolution != ResolutionSyntaxOnly || report.ProcessCoordinate != wantCoordinate {
		return fmt.Errorf("unknown catalog resolution is not exact")
	}
	if report.SourceLines != 5 || !validDigest(report.SourceDigest) || report.Diagnostic == "" {
		return fmt.Errorf("unknown catalog source evidence is invalid")
	}
	wantMetrics := OperationSpecMetrics{
		MetricID: OperationSpecMetricID, FixedAxisTotal: OperationSpecAxisTotal,
		UnknownPathCount: 1, OpenClaims: OperationSpecAxisTotal,
		TransitionEventTotal: OperationClaimEventTotal, RegistrationEventTotal: OperationSpecAxisTotal,
		EvidenceUnavailableTotal: OperationSpecAxisTotal,
	}
	if report.OperationSpecMetrics != wantMetrics {
		return fmt.Errorf("unknown catalog metrics are not exact")
	}
	for _, claim := range report.Claims {
		if claim.Status != ClaimStatusOpen || claim.EvidenceDigest != "" {
			return fmt.Errorf("unknown claim %s was hidden", claim.ClaimID)
		}
	}
	if err := validateClaimTransitionLedger(report); err != nil {
		return err
	}
	if len(report.OperationSpecs) != 0 || len(report.Indicators) != 0 || len(report.Views) != 0 || len(report.Proofs) != 0 {
		return fmt.Errorf("unknown catalog overclaimed derived evidence")
	}
	if report.Authority != (Authority{}) || len(report.NonClaims) != 5 || report.Digest != reportDigest(report) {
		return fmt.Errorf("unknown catalog boundary changed")
	}
	return nil
}
