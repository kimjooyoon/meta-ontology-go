package generation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const OperationObservationBundleSchema = "gooo/meta-operation-observation-bundle/v2"

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

func (bundle *OperationObservationBundle) UnmarshalJSON(data []byte) error {
	type wire OperationObservationBundle
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var candidate wire
	if err := decoder.Decode(&candidate); err != nil {
		return fmt.Errorf("decode operation observation bundle: %w", err)
	}
	if err := ensureIndicatorLedgerEOF(decoder); err != nil {
		return fmt.Errorf("decode operation observation bundle: %w", err)
	}
	if candidate.Schema != OperationObservationBundleSchema {
		return fmt.Errorf("unsupported operation observation bundle schema %q", candidate.Schema)
	}
	if candidate.Receipts == nil {
		candidate.Receipts = []OperationReceipt{}
	}
	if candidate.Failures == nil {
		candidate.Failures = []ObservationFailure{}
	}
	decoded := OperationObservationBundle(candidate)
	decoded.Receipts = normalizeOperationReceipts(decoded.Receipts)
	decoded.Failures = normalizeObservationFailures(decoded.Failures)
	if !reflect.DeepEqual(decoded, SealObservationBundle(decoded)) {
		return fmt.Errorf("operation observation bundle canonical replay mismatch")
	}
	*bundle = decoded
	return nil
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
	bundle.ReplayDigest = digestPair(bundle.PlanDigest, operationObservationReplayDigest(bundle))
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
		bundle.ReplayDigest != digestPair(bundle.PlanDigest, operationObservationReplayDigest(bundle)) {
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
		if !actionExists || !receiptMatchesAction(plan, action, receipt) {
			return fmt.Errorf("operation observation receipt binding mismatch for %s", receipt.ActionIndicatorID)
		}
		if evidence.Schema != OperationInstanceEvidenceSchema ||
			evidence.ActionIndicatorID != receipt.ActionIndicatorID ||
			evidence.Subject != action.Subject ||
			evidence.HeadSHA != plan.HeadSHA ||
			evidence.Subject == "" ||
			evidence.OperationID == "" ||
			!validEvidenceDigest(evidence.ContractEvidenceDigest) ||
			!validEvidenceDigest(evidence.InstanceEvidenceDigest) ||
			!validEvidenceOrigin(*evidence) ||
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
			!validObservationFailureEvidence(failure.FailureEvidence) ||
			!validCounterexampleRelations(failure.DerivedRelations) {
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
		ReceiptUnknownClassUnexpectedEvidence, ReceiptUnknownClassDependencyBlocked,
		"STALE", "AMBIGUOUS", "UNBOUNDED":
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
			!validObservationFailureEvidence(failure.FailureEvidence) ||
			!validCounterexampleRelations(failure.DerivedRelations) {
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
	return validCanonicalCommand(observation.Command) && observation.StdoutBytes >= 0 &&
		observation.StderrBytes >= 0 && validEvidenceDigest(observation.RawStdoutDigest) &&
		validEvidenceDigest(observation.StdoutDigest) && validEvidenceDigest(observation.RawStderrDigest) &&
		validEvidenceDigest(observation.StderrDigest)
}

const EvidenceOriginInputReceipt = "INPUT_RECEIPT"

func validEvidenceOrigin(evidence OperationInstanceEvidence) bool {
	if evidence.EvidenceOrigin == "" && evidence.SourceReceiptDigest == "" {
		return true
	}
	return evidence.EvidenceOrigin == EvidenceOriginInputReceipt &&
		validEvidenceDigest(evidence.SourceReceiptDigest)
}

func validCanonicalCommand(command []string) bool {
	if len(command) == 0 {
		return false
	}
	for _, argument := range command {
		if argument == "" || absoluteCommandArgument(argument) {
			return false
		}
	}
	return true
}

func absoluteCommandArgument(argument string) bool {
	if filepath.IsAbs(argument) || strings.HasPrefix(argument, "//") || strings.HasPrefix(argument, `\\`) {
		return true
	}
	if len(argument) >= 3 && argument[1] == ':' &&
		((argument[0] >= 'a' && argument[0] <= 'z') || (argument[0] >= 'A' && argument[0] <= 'Z')) &&
		(argument[2] == '/' || argument[2] == '\\') {
		return true
	}
	if _, after, ok := strings.Cut(argument, "="); ok {
		return absoluteCommandArgument(after)
	}
	return false
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

func validCounterexampleRelations(relations []CounterexampleRelation) bool {
	seen := make(map[string]bool, len(relations))
	for _, relation := range relations {
		if relation.Counterexample == "" || relation.DerivedFrom == "" ||
			relation.Relation != "DERIVED_FROM" || seen[relation.Counterexample] {
			return false
		}
		seen[relation.Counterexample] = true
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
		result[index].DerivedRelations = normalizeCounterexampleRelations(result[index].DerivedRelations)
		result[index].Diagnostics = append([]string(nil), result[index].Diagnostics...)
	}
	sort.SliceStable(result, func(left, right int) bool {
		leftKey, _ := json.Marshal(result[left])
		rightKey, _ := json.Marshal(result[right])
		return string(leftKey) < string(rightKey)
	})
	return result
}

func normalizeCounterexampleRelations(relations []CounterexampleRelation) []CounterexampleRelation {
	result := append([]CounterexampleRelation{}, relations...)
	sort.Slice(result, func(left, right int) bool {
		return result[left].Counterexample < result[right].Counterexample
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

type replayProcessProjection struct {
	Command      []string `json:"command"`
	ExitCode     int      `json:"exit_code"`
	StdoutBytes  int      `json:"stdout_bytes"`
	StdoutDigest string   `json:"stdout_digest"`
	StderrBytes  int      `json:"stderr_bytes"`
	StderrDigest string   `json:"stderr_digest"`
}

type replayReceiptProjection struct {
	ActionIndicatorID      string                   `json:"action_indicator_id"`
	Subject                string                   `json:"subject"`
	HeadSHA                string                   `json:"head_sha"`
	OperationID            string                   `json:"operation_id"`
	ContractEvidenceDigest string                   `json:"contract_evidence_digest"`
	InstanceEvidenceDigest string                   `json:"instance_evidence_digest"`
	EvidenceOrigin         string                   `json:"evidence_origin,omitempty"`
	SourceReceiptDigest    string                   `json:"source_receipt_digest,omitempty"`
	Indicators             []IndicatorReceipt       `json:"indicators"`
	Executor               replayProcessProjection  `json:"executor"`
	Evaluator              replayProcessProjection  `json:"evaluator"`
	Verifier               *replayProcessProjection `json:"verifier,omitempty"`
}

type replayFailureProjection struct {
	ActionIndicatorID string                       `json:"action_indicator_id"`
	Decision          string                       `json:"decision"`
	Stage             string                       `json:"stage"`
	Step              string                       `json:"step"`
	Reason            string                       `json:"reason"`
	UnknownClass      string                       `json:"unknown_class,omitempty"`
	NextOperation     string                       `json:"next_operation"`
	BlockedBy         []string                     `json:"blocked_by"`
	FailureEvidence   []ObservationFailureEvidence `json:"failure_evidence,omitempty"`
	Counterexample    string                       `json:"counterexample,omitempty"`
	DerivedRelations  []CounterexampleRelation     `json:"derived_relations,omitempty"`
	Diagnostics       []string                     `json:"diagnostics,omitempty"`
	Executor          replayProcessProjection      `json:"executor"`
}

func operationObservationReplayDigest(bundle OperationObservationBundle) string {
	receipts := make([]replayReceiptProjection, 0, len(bundle.Receipts))
	for _, receipt := range bundle.Receipts {
		if receipt.InstanceEvidence == nil {
			continue
		}
		evidence := receipt.InstanceEvidence
		var verifier *replayProcessProjection
		if evidence.VerifierObservation != nil {
			value := replayProcess(*evidence.VerifierObservation)
			verifier = &value
		}
		receipts = append(receipts, replayReceiptProjection{
			ActionIndicatorID: receipt.ActionIndicatorID, Subject: evidence.Subject,
			HeadSHA: evidence.HeadSHA, OperationID: evidence.OperationID,
			ContractEvidenceDigest: evidence.ContractEvidenceDigest,
			InstanceEvidenceDigest: evidence.InstanceEvidenceDigest,
			EvidenceOrigin:         evidence.EvidenceOrigin, SourceReceiptDigest: evidence.SourceReceiptDigest,
			Indicators: receipt.Indicators,
			Executor:   replayProcess(evidence.ExecutorObservation),
			Evaluator:  replayProcess(evidence.EvaluatorObservation), Verifier: verifier,
		})
	}
	failures := make([]replayFailureProjection, 0, len(bundle.Failures))
	for _, failure := range bundle.Failures {
		failures = append(failures, replayFailureProjection{
			ActionIndicatorID: failure.ActionIndicatorID, Decision: failure.Decision,
			Stage: failure.Stage, Step: failure.Step, Reason: failure.Reason,
			UnknownClass: failure.UnknownClass, NextOperation: failure.NextOperation,
			BlockedBy: failure.BlockedBy, FailureEvidence: failure.FailureEvidence,
			Counterexample: failure.Counterexample, DerivedRelations: failure.DerivedRelations,
			Diagnostics: failure.Diagnostics, Executor: replayProcess(failure.Executor),
		})
	}
	return digestJSON(struct {
		Schema            string                    `json:"schema"`
		BaseSHA           string                    `json:"base_sha"`
		HeadSHA           string                    `json:"head_sha"`
		PlanDigest        string                    `json:"plan_digest"`
		ManifestDigest    string                    `json:"manifest_digest"`
		Receipts          []replayReceiptProjection `json:"receipts"`
		Failures          []replayFailureProjection `json:"failures"`
		ObservationTotal  int                       `json:"observation_total"`
		ReplayComparisons int                       `json:"replay_comparisons"`
	}{Schema: bundle.Schema, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA,
		PlanDigest: bundle.PlanDigest, ManifestDigest: bundle.ManifestDigest,
		Receipts: receipts, Failures: failures, ObservationTotal: bundle.ObservationTotal,
		ReplayComparisons: bundle.ReplayComparisons})
}

func replayProcess(observation ProcessObservation) replayProcessProjection {
	return replayProcessProjection{Command: append([]string{}, observation.Command...), ExitCode: observation.ExitCode,
		StdoutBytes:  observation.StdoutBytes,
		StdoutDigest: observation.StdoutDigest, StderrBytes: observation.StderrBytes,
		StderrDigest: observation.StderrDigest}
}
