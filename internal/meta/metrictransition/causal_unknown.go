package metrictransition

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

func validateCausalUnknowns(ledger transformationeffect.Ledger, report generation.ReceiptReport) error {
	projection, err := deriveCausalUnknowns(report)
	if err != nil {
		return err
	}
	if ledger.DirectUnknownCount != projection.DirectUnknownCount ||
		ledger.DependencyBlockedUnknownCount != projection.DependencyBlockedUnknownCount ||
		ledger.UnknownCausalDigest != projection.Digest {
		return fmt.Errorf("metric transition causal unknown projection diverged")
	}
	if ledger.OperationOutcome == effectOutcomeMixedRefuted &&
		(projection.DirectUnknownCount != 0 || projection.DependencyBlockedUnknownCount != len(report.Unknowns)) {
		return fmt.Errorf("metric transition mixed outcome has unbound unknowns")
	}
	return nil
}
