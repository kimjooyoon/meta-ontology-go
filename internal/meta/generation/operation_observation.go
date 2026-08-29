package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const OperationObservationBundleSchema = "gooo/meta-operation-observation-bundle/v1"

type OperationObservationBundle struct {
	Schema            string               `json:"schema"`
	BaseSHA           string               `json:"base_sha"`
	HeadSHA           string               `json:"head_sha"`
	PlanDigest        string               `json:"plan_digest"`
	ManifestDigest    string               `json:"manifest_digest"`
	Receipts          []OperationReceipt   `json:"receipts"`
	Failures          []ObservationFailure `json:"failures"`
	ObservationTotal  int                  `json:"observation_total"`
	ReplayComparisons int                  `json:"replay_comparisons"`
	BundleDigest      string               `json:"bundle_digest"`
	ReplayDigest      string               `json:"replay_digest"`
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
	bundle.Failures = normalizeObservationFailures(bundle.Failures)
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
		if !validIndicatorObservations(receipt, action, *evidence) {
			return fmt.Errorf("operation indicator observations are incomplete for %s", receipt.ActionIndicatorID)
		}
		if seen[receipt.ActionIndicatorID] {
			return fmt.Errorf("operation observation duplicate for %s", receipt.ActionIndicatorID)
		}
		seen[receipt.ActionIndicatorID] = true
	}
	for _, failure := range bundle.Failures {
		if failure.ActionIndicatorID == "" || seen[failure.ActionIndicatorID] ||
			!actionsExist(actions, failure.ActionIndicatorID) ||
			failure.Decision == "" || failure.Stage == "" || failure.Step == "" || failure.Reason == "" ||
			!validObservationFailureDecision(failure) || failure.NextOperation == "" ||
			failure.BlockedBy == nil || !validProcessObservation(failure.Executor) ||
			!validObservationFailureEvidence(failure.FailureEvidence) {
			return fmt.Errorf("invalid operation observation failure")
		}
		seen[failure.ActionIndicatorID] = true
	}
	if len(seen) != len(plan.Selected) {
		return fmt.Errorf("operation observation subject coverage mismatch")
	}
	return nil
}

func validObservationFailureDecision(failure ObservationFailure) bool {
	if failure.Decision == "REFUTED" {
		return failure.UnknownClass == ""
	}
	if failure.Decision != "UNKNOWN" {
		return false
	}
	switch failure.UnknownClass {
	case ReceiptUnknownClassDirectMissing, ReceiptUnknownClassMalformedEvidence,
		ReceiptUnknownClassUnexpectedEvidence, ReceiptUnknownClassDependencyBlocked:
		return true
	default:
		return false
	}
}

func validReceiptFailureList(failures []ObservationFailure) bool {
	canonical := normalizeObservationFailures(failures)
	if len(canonical) != len(failures) {
		return false
	}
	for index, failure := range failures {
		if failure.ActionIndicatorID == "" || failure.Stage == "" || failure.Step == "" ||
			failure.Reason == "" || failure.NextOperation == "" || failure.BlockedBy == nil ||
			!validObservationFailureDecision(failure) || !validProcessObservation(failure.Executor) ||
			!validObservationFailureEvidence(failure.FailureEvidence) {
			return false
		}
		left, _ := json.Marshal(canonical[index])
		right, _ := json.Marshal(failure)
		if string(left) != string(right) {
			return false
		}
	}
	return true
}

func validIndicatorObservations(receipt OperationReceipt, action Action, evidence OperationInstanceEvidence) bool {
	indicators, valid := indicatorReceiptIndex(receipt.Indicators)
	if !valid || len(indicators) != len(action.RequiredIndicatorIDs) {
		return false
	}
	for _, identifier := range action.RequiredIndicatorIDs {
		observation, exists := indicators[identifier]
		if !exists || observation.Observation == nil ||
			observation.Observation.Schema != IndicatorObservationSchema ||
			observation.Observation.IndicatorID != identifier ||
			observation.Observation.Subject != action.Subject ||
			observation.Observation.HeadSHA != evidence.HeadSHA ||
			observation.Observation.OperationID != evidence.OperationID ||
			observation.Observation.ValueKind != "integer" ||
			observation.Observation.ActualValue < 0 || observation.Observation.ActualValue > 1 ||
			observation.Observation.ExpectedPredicate != "equal" ||
			observation.Observation.ExpectedBound != 1 ||
			(observation.Verdict == IndicatorVerdictPass && observation.Observation.ActualValue != 1) ||
			(observation.Verdict == IndicatorVerdictFail && observation.Observation.ActualValue == 1) ||
			(observation.Verdict == IndicatorVerdictUnknown && observation.Observation.ActualValue != 0) ||
			observation.Observation.TransformedSubject == "" ||
			observation.EvidenceDigest != digestJSON(observation.Observation) {
			return false
		}
		if action.Operation == sourcepolicy.OperationExtractFunction &&
			!validExtractFunctionObservation(*observation.Observation, action) {
			return false
		}
	}
	return true
}

func validExtractFunctionObservation(observation IndicatorObservation, action Action) bool {
	subject, err := sourcepolicy.ParseSourceSubject(action.Subject)
	if err != nil {
		return false
	}
	if action.SourceIndicator.Value <= action.SourceIndicator.Limit ||
		observation.BeforeFunctionLines != action.SourceIndicator.Value ||
		observation.BeforeFunctionLines <= action.SourceIndicator.Limit ||
		observation.AfterFunctionLines <= 0 ||
		observation.AfterFunctionLines > action.SourceIndicator.Limit ||
		observation.AfterFunctionLines >= observation.BeforeFunctionLines {
		return false
	}
	prefix := subject.Path + "#" + subject.Name + "=>"
	return strings.HasPrefix(observation.TransformedSubject, prefix)
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

func validObservationFailureEvidence(evidence []ObservationFailureEvidence) bool {
	seen := make(map[string]bool, len(evidence))
	for _, item := range evidence {
		if item.IndicatorID == "" || seen[item.IndicatorID] || item.Counterexample == "" ||
			item.Observed < 0 || item.Expected < 0 || item.Decision == "PASS" || item.Decision == "" {
			return false
		}
		seen[item.IndicatorID] = true
	}
	return true
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
	bundle.Failures = normalizeObservationFailures(bundle.Failures)
	sort.Slice(bundle.Receipts, func(left, right int) bool {
		return bundle.Receipts[left].ActionIndicatorID < bundle.Receipts[right].ActionIndicatorID
	})
	return bundle
}

func normalizeObservationFailures(failures []ObservationFailure) []ObservationFailure {
	result := append([]ObservationFailure{}, failures...)
	for index := range result {
		result[index].FailureEvidence = normalizeObservationFailureEvidence(result[index].FailureEvidence)
	}
	sort.SliceStable(result, func(left, right int) bool {
		leftKey, _ := json.Marshal(result[left])
		rightKey, _ := json.Marshal(result[right])
		return string(leftKey) < string(rightKey)
	})
	return result
}

func normalizeObservationFailureEvidence(evidence []ObservationFailureEvidence) []ObservationFailureEvidence {
	result := append([]ObservationFailureEvidence{}, evidence...)
	sort.Slice(result, func(left, right int) bool {
		return result[left].IndicatorID < result[right].IndicatorID
	})
	return result
}
