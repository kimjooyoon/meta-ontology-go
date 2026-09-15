package metriccounterfactualverify

import (
	"fmt"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactual"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func validateLedger(ledger metric.Ledger) error {
	if ledger.Schema != metric.LedgerSchema || !metric.ValidLedger(ledger) {
		return fmt.Errorf("invalid ledger")
	}
	if !metric.ValidManifest(ledger.Manifest) || !metric.ValidPlan(ledger.Plan) ||
		!metric.ValidState(ledger.Before) || !metric.ValidState(ledger.After) {
		return fmt.Errorf("invalid sealed input")
	}
	if ledger.PromotionAuthorized || ledger.RepositoryWorkspaceWrites ||
		ledger.ExecutionPolicy != "DISPOSABLE_TEMP_ROOT_ONLY" {
		return fmt.Errorf("unsafe ledger policy")
	}
	receiptSetDigest, err := artifact.Digest(ledger.Receipts)
	if err != nil {
		return err
	}
	if ledger.Evidence.ManifestDigest != ledger.Manifest.Digest ||
		ledger.Evidence.PlanDigest != ledger.Plan.Digest ||
		ledger.Evidence.ReceiptSetDigest != receiptSetDigest ||
		ledger.Evidence.BeforeDigest != ledger.Before.Digest ||
		ledger.Evidence.AfterDigest != ledger.After.Digest {
		return fmt.Errorf("unbound ledger evidence")
	}
	return nil
}
