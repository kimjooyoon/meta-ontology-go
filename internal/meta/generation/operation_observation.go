package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

const OperationObservationBundleSchema = "gooo/meta-operation-observation-bundle/v1"

type OperationObservationBundle struct {
	Schema              string             `json:"schema"`
	BaseSHA             string             `json:"base_sha"`
	HeadSHA             string             `json:"head_sha"`
	PlanDigest          string             `json:"plan_digest"`
	ManifestDigest      string             `json:"manifest_digest"`
	Receipts            []OperationReceipt `json:"receipts"`
	Failures            []ObservationFailure `json:"failures"`
	ObservationTotal    int                `json:"observation_total"`
	ReplayComparisons   int                `json:"replay_comparisons"`
	BundleDigest        string             `json:"bundle_digest"`
	ReplayDigest        string             `json:"replay_digest"`
}

func AttachInstanceEvidence(receipt OperationReceipt, evidence OperationInstanceEvidence) OperationReceipt {
	receipt.InstanceEvidence = &evidence
	receipt.ReceiptDigest = ""
	receipt.ReceiptDigest = operationReceiptDigest(receipt)
	return receipt
}

func SealObservationBundle(bundle OperationObservationBundle) OperationObservationBundle {
	bundle.Schema = OperationObservationBundleSchema
	bundle.Receipts = normalizeOperationReceipts(bundle.Receipts)
	bundle.BundleDigest, bundle.ReplayDigest = "", ""
	bundle.BundleDigest = operationObservationBundleDigest(bundle)
	bundle.ReplayDigest = digestPair(bundle.PlanDigest, bundle.BundleDigest)
	return bundle
}

func EncodeObservationBundle(bundle OperationObservationBundle) ([]byte, error) {
	payload, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func ValidateObservationBundle(bundle OperationObservationBundle, plan Plan, manifest ExecutionManifest) error {
	if bundle.Schema != OperationObservationBundleSchema ||
		bundle.BaseSHA != plan.BaseSHA || bundle.HeadSHA != plan.HeadSHA ||
		bundle.PlanDigest != plan.PlanDigest ||
		bundle.ManifestDigest != manifest.ManifestDigest ||
		!validSHA(bundle.BaseSHA) || !validSHA(bundle.HeadSHA) ||
		!validDigest(bundle.BundleDigest) ||
		bundle.BundleDigest != operationObservationBundleDigest(bundle) ||
		bundle.ReplayDigest != digestPair(bundle.PlanDigest, bundle.BundleDigest) {
		return fmt.Errorf("operation observation bundle context or digest mismatch")
	}
	if len(bundle.Receipts)+len(bundle.Failures) != len(plan.Selected) ||
		bundle.ObservationTotal != len(bundle.Receipts)+len(bundle.Failures) ||
		bundle.ReplayComparisons < len(bundle.Receipts) {
		return fmt.Errorf("operation observation bundle coverage mismatch")
	}
	seen := make(map[string]bool, len(bundle.Receipts)+len(bundle.Failures))
	actions := make(map[string]Action, len(plan.Selected))
	for _, action := range plan.Selected {
		actions[action.IndicatorID] = action
	}
	for _, receipt := range bundle.Receipts {
		if receipt.InstanceEvidence == nil {
			return fmt.Errorf("operation observation is missing instance evidence")
		}
		evidence := receipt.InstanceEvidence
		action, actionExists := actions[receipt.ActionIndicatorID]
		if evidence.Schema != OperationInstanceEvidenceSchema ||
			evidence.ActionIndicatorID != receipt.ActionIndicatorID ||
			!actionExists || evidence.Subject != action.Subject ||
			evidence.HeadSHA != plan.HeadSHA ||
			evidence.Subject == "" ||
			evidence.OperationID == "" ||
			!validEvidenceDigest(evidence.ContractEvidenceDigest) ||
			!validEvidenceDigest(evidence.InstanceEvidenceDigest) ||
			evidence.ReplayComparisons < 1 || !evidence.ReplayMatch ||
			evidence.ExecutorObservation.ExitCode != 0 ||
			evidence.EvaluatorObservation.ExitCode != 0 ||
			evidence.VerifierObservation == nil || evidence.VerifierObservation.ExitCode != 0 ||
			!validProcessObservation(evidence.ExecutorObservation) ||
			!validProcessObservation(evidence.EvaluatorObservation) ||
			!validProcessObservation(*evidence.VerifierObservation) {
			return fmt.Errorf("operation observation is incomplete for %s", receipt.ActionIndicatorID)
		}
		if seen[receipt.ActionIndicatorID] {
			return fmt.Errorf("operation observation duplicate for %s", receipt.ActionIndicatorID)
		}
		seen[receipt.ActionIndicatorID] = true
	}
	for _, failure := range bundle.Failures {
		if failure.ActionIndicatorID == "" || seen[failure.ActionIndicatorID] ||
			!actionsExist(actions, failure.ActionIndicatorID) ||
			failure.Stage == "" || failure.Step == "" || failure.Reason == "" ||
			failure.UnknownClass == "" || failure.NextOperation == "" ||
			failure.BlockedBy == nil || !validProcessObservation(failure.Executor) {
			return fmt.Errorf("invalid operation observation failure")
		}
		seen[failure.ActionIndicatorID] = true
	}
	if len(seen) != len(plan.Selected) {
		return fmt.Errorf("operation observation subject coverage mismatch")
	}
	return nil
}

func actionsExist(actions map[string]Action, identifier string) bool {
	_, ok := actions[identifier]
	return ok
}

func validProcessObservation(observation ProcessObservation) bool {
	return len(observation.Command) != 0 && observation.StdoutBytes >= 0 &&
		observation.StderrBytes >= 0 && validEvidenceDigest(observation.StdoutDigest) &&
		validEvidenceDigest(observation.StderrDigest)
}

func validEvidenceDigest(value string) bool {
	return len(value) == len("sha256:")+64 && value[:len("sha256:")] == "sha256:" &&
		validDigest(value[len("sha256:"):])
}

func operationObservationBundleDigest(bundle OperationObservationBundle) string {
	unsigned := bundle
	unsigned.BundleDigest, unsigned.ReplayDigest = "", ""
	payload, _ := json.Marshal(unsigned)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func NormalizeObservationBundle(bundle OperationObservationBundle) OperationObservationBundle {
	bundle.Receipts = normalizeOperationReceipts(bundle.Receipts)
	sort.Slice(bundle.Receipts, func(left, right int) bool {
		return bundle.Receipts[left].ActionIndicatorID < bundle.Receipts[right].ActionIndicatorID
	})
	return bundle
}
