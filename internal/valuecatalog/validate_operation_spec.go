package valuecatalog

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func validateOperationSpecEvidence(report Report) error {
	if len(report.OperationSpecs) != 1 || valueexecution.ValidateOperationSpec(report.OperationSpecs[0]) != nil {
		return fmt.Errorf("operation spec catalog is not exact")
	}
	metrics := report.OperationSpecMetrics
	if metrics.MetricID != OperationSpecMetricID || metrics.FixedAxisTotal != OperationSpecAxisTotal ||
		metrics.VerifiedTotal != OperationSpecAxisTotal || metrics.CoverageBasisPoints != 10_000 ||
		metrics.UnknownPathCount != 0 || metrics.OpenClaims != 0 || metrics.DischargedClaims != OperationSpecAxisTotal {
		return fmt.Errorf("operation spec OS9 metric is not exact")
	}
	if len(report.Claims) != OperationSpecAxisTotal {
		return fmt.Errorf("operation spec claim denominator changed")
	}
	for _, claim := range report.Claims {
		if claim.Status != "DISCHARGED" || !validDigest(claim.EvidenceDigest) {
			return fmt.Errorf("operation spec claim %s is not discharged", claim.ClaimID)
		}
	}
	if report.ProcessCoordinate.Stage != "REDUCE" || report.ProcessCoordinate.Step != "close-os9" || report.ProcessCoordinate.Reason != report.Reason {
		return fmt.Errorf("operation spec process coordinate is not closed")
	}
	for index, indicator := range report.Indicators {
		want := CatalogMetricID
		if index >= CatalogIndicatorCount-OperationSpecAxisTotal {
			want = OperationSpecMetricID
		}
		if indicator.MetricID != want {
			return fmt.Errorf("indicator %s metric binding changed", indicator.ID)
		}
	}
	return nil
}
