package metriccounterfactual

import (
	"fmt"
	"os"
	"strings"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func Generate(repository, subjectSHA string) (Ledger, error) {
	if !ValidSubject(repository, subjectSHA) {
		return Ledger{}, fmt.Errorf("invalid repository subject")
	}
	manifest, err := BaselineManifest()
	if err != nil {
		return Ledger{}, err
	}
	plan, err := CounterfactualPlan()
	if err != nil {
		return Ledger{}, err
	}
	root, err := os.MkdirTemp("", "gooo-metric-counterfactual-")
	if err != nil {
		return Ledger{}, err
	}
	defer os.RemoveAll(root)
	if err := Materialize(root, manifest); err != nil {
		return Ledger{}, err
	}
	before, err := Measure(root)
	if err != nil {
		return Ledger{}, err
	}
	receipts, err := ApplyPlan(root, plan)
	if err != nil {
		return Ledger{}, err
	}
	after, err := Measure(root)
	if err != nil {
		return Ledger{}, err
	}
	delta := ComputeDelta(before, after)
	indicators, err := EvaluateIndicators(manifest, plan, before, after, receipts, delta)
	if err != nil || !AllSatisfied(indicators) {
		return Ledger{}, fmt.Errorf("counterfactual indicators are unsatisfied: %w", err)
	}
	receiptSetDigest, err := artifact.Digest(receipts)
	if err != nil {
		return Ledger{}, err
	}
	return SealLedger(Ledger{
		Schema: LedgerSchema, Repository: repository, SubjectSHA: subjectSHA,
		ExecutionPolicy: "DISPOSABLE_TEMP_ROOT_ONLY", RepositoryWorkspaceWrites: false,
		Manifest: manifest, Plan: plan, Receipts: receipts, Before: before, After: after, Delta: delta,
		Evidence: Evidence{ManifestDigest: manifest.Digest, PlanDigest: plan.Digest,
			ReceiptSetDigest: receiptSetDigest, BeforeDigest: before.Digest, AfterDigest: after.Digest},
		Indicators: indicators, PromotionAuthorized: false,
	})
}

func SealLedger(value Ledger) (Ledger, error) {
	value.Digest = ""
	digest, err := artifact.Digest(value)
	value.Digest = digest
	return value, err
}

func ValidLedger(value Ledger) bool {
	digest := value.Digest
	sealed, err := SealLedger(value)
	return err == nil && digest == sealed.Digest
}

func ValidSubject(repository, subjectSHA string) bool {
	if strings.TrimSpace(repository) == "" || len(subjectSHA) != 40 {
		return false
	}
	for _, char := range subjectSHA {
		if !strings.ContainsRune("0123456789abcdef", char) {
			return false
		}
	}
	return true
}
