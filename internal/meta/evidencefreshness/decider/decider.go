package decider

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
)

type axisSpec struct {
	name  string
	stage string
	get   func(model.EvidenceTuple) string
}

var axes = []axisSpec{
	{name: "subject", stage: model.StageSubject, get: func(tuple model.EvidenceTuple) string { return tuple.Subject }},
	{name: "material", stage: model.StageMaterial, get: func(tuple model.EvidenceTuple) string { return tuple.Material }},
	{name: "recipe", stage: model.StageRecipe, get: func(tuple model.EvidenceTuple) string { return tuple.Recipe }},
	{name: "environment", stage: model.StageEnvironment, get: func(tuple model.EvidenceTuple) string { return tuple.Environment }},
	{name: "runner", stage: model.StageRunner, get: func(tuple model.EvidenceTuple) string { return tuple.Runner }},
	{name: "verifier", stage: model.StageVerifier, get: func(tuple model.EvidenceTuple) string { return tuple.Verifier }},
}

// Decide consumes only a receipt and a current context. It does not import the
// source observer, compiler parser, or producer package.
func Decide(receiptRaw, contextRaw []byte) model.Verdict {
	receipt, receiptErr := model.DecodeStrict[model.Receipt](receiptRaw)
	context, contextErr := model.DecodeStrict[model.Context](contextRaw)
	if receiptErr != nil || contextErr != nil {
		return finish(unknown("SUBJECT_UNKNOWN", model.StageSubject, "decode-receipt-or-context", "receipt-or-context"), receipt, context)
	}
	if receipt.Schema != model.ReceiptSchema || !model.ValidHead(receipt.HeadSHA) || !model.VerifyReceiptDigest(receipt) {
		return finish(unknown("RECEIPT_UNKNOWN", model.StageSubject, "verify-receipt", "receipt-identity"), receipt, context)
	}
	if context.Schema != model.ContextSchema || context.Consumer != model.ConsumerID {
		return finish(unknown("CONSUMER_UNKNOWN", model.StageVerifier, "verify-context", "consumer-identity"), receipt, context)
	}
	if receipt.Producer != model.ProducerID || receipt.Consumer != model.ConsumerID || receipt.MetaOperation != model.MetaOperationID || receipt.ProofChoice == "" {
		return finish(refuted("RECEIPT_META_BINDING_INVALID", model.StageVerifier, "verify-receipt-meta", "meta-binding"), receipt, context)
	}
	if receipt.RepositoryWrites != 0 || receipt.MutationAuthority {
		return finish(refuted("READ_ONLY_EFFECT_INVALID", model.StageVerifier, "verify-effects", "read-only-effects"), receipt, context)
	}

	var changed []string
	var firstChanged *axisSpec
	for index := range axes {
		axis := axes[index]
		receiptValue, currentValue := axis.get(receipt.Tuple), axis.get(context.Tuple)
		if receiptValue == "" || currentValue == "" {
			if firstChanged != nil {
				return finish(stale(axisReason(*firstChanged), firstChanged.stage, axisStep(*firstChanged), changed), receipt, context)
			}
			return finish(unknown(strings.ToUpper(axis.name)+"_UNKNOWN", axis.stage, "read-"+axis.name, axis.name+"-identity"), receipt, context)
		}
		if receiptValue != currentValue {
			changed = append(changed, axis.name)
			if firstChanged == nil {
				firstChanged = &axis
			}
		}
	}
	if firstChanged != nil {
		return finish(stale(axisReason(*firstChanged), firstChanged.stage, axisStep(*firstChanged), changed), receipt, context)
	}
	if receipt.Boundary.EnvironmentBoundary == "" || context.EnvironmentBoundary == "" {
		return finish(unknown("ENVIRONMENT_BOUNDARY_UNKNOWN", model.StageEnvironment, "read-environment-boundary", "environment-boundary"), receipt, context)
	}
	if receipt.Boundary.EnvironmentBoundary != context.EnvironmentBoundary {
		return finish(stale("ENVIRONMENT_BOUNDARY_CHANGED", model.StageEnvironment, "compare-environment-boundary", []string{"environment-boundary"}), receipt, context)
	}
	if receipt.Boundary.ObservationEpoch <= 0 || receipt.Boundary.ValidThroughEpoch < receipt.Boundary.ObservationEpoch || context.CurrentEpoch <= 0 {
		return finish(unknown("TEMPORAL_BOUNDARY_UNKNOWN", model.StageVerifier, "read-validity-boundary", "temporal-boundary"), receipt, context)
	}
	if context.CurrentEpoch < receipt.Boundary.ObservationEpoch {
		return finish(unknown("TEMPORAL_BOUNDARY_PRECEDES_OBSERVATION", model.StageVerifier, "compare-validity-boundary", "temporal-boundary"), receipt, context)
	}
	if context.CurrentEpoch > receipt.Boundary.ValidThroughEpoch {
		return finish(stale("TEMPORAL_BOUNDARY_EXPIRED", model.StageVerifier, "check-validity-boundary", []string{"temporal-boundary"}), receipt, context)
	}
	return finish(model.Verdict{Schema: model.VerdictSchema, State: model.StateFresh,
		Decision: model.DecisionPass, Resolution: model.ResolutionExact,
		Reason:     "TUPLE_EXACT_AND_BOUNDARY_CURRENT",
		Coordinate: model.Coordinate{Stage: model.StageVerifier, Step: "accept-current-evidence", Reason: "TUPLE_EXACT_AND_BOUNDARY_CURRENT"}}, receipt, context)
}

func axisReason(axis axisSpec) string {
	return strings.ToUpper(axis.name) + "_CHANGED"
}

func axisStep(axis axisSpec) string {
	return "compare-" + axis.name
}

func stale(reason, stage, step string, dimensions []string) model.Verdict {
	return model.Verdict{Schema: model.VerdictSchema, State: model.StateStale,
		Decision: model.DecisionFailClosed, Resolution: model.ResolutionInvariant, Reason: reason,
		Coordinate: model.Coordinate{Stage: stage, Step: step, Reason: reason}, ChangedDimensions: dimensions}
}

func unknown(reason, stage, step, dimension string) model.Verdict {
	return model.Verdict{Schema: model.VerdictSchema, State: model.StateUnknown,
		Decision: model.DecisionFailClosed, Resolution: model.ResolutionLower, Reason: reason,
		Coordinate: model.Coordinate{Stage: stage, Step: step, Reason: reason}, ChangedDimensions: []string{dimension}}
}

func refuted(reason, stage, step, dimension string) model.Verdict {
	return model.Verdict{Schema: model.VerdictSchema, State: "REFUTED",
		Decision: model.DecisionFailClosed, Resolution: model.ResolutionInvariant, Reason: reason,
		Coordinate: model.Coordinate{Stage: stage, Step: step, Reason: reason}, ChangedDimensions: []string{dimension}}
}

func finish(verdict model.Verdict, receipt model.Receipt, context model.Context) model.Verdict {
	verdict.ReceiptDigest = receipt.Digest
	verdict.ContextDigest = model.DigestJSON(context)
	verdict.Transition = transition(verdict, receipt.Digest)
	verdict.Checks = checks(verdict, receipt)
	return model.SealVerdict(verdict)
}

func transition(verdict model.Verdict, receiptDigest string) model.ClaimTransition {
	from, to, preservation := "CLAIM_JUSTIFIED", "CLAIM_UNKNOWN", "DO_NOT_PRESERVE"
	switch verdict.State {
	case model.StateFresh:
		to, preservation = "CLAIM_PRESERVED", "PRESERVE_EXACT"
	case model.StateStale:
		to = "CLAIM_STALE"
	}
	return model.ClaimTransition{ClaimID: "gooo://evidence-freshness/claim/checked-source", From: from,
		To: to, Preservation: preservation, Coordinate: verdict.Coordinate, EvidenceDigest: receiptDigest}
}

func checks(verdict model.Verdict, receipt model.Receipt) []model.CheckResult {
	status := "PASS"
	if verdict.State == model.StateUnknown || verdict.State == "REFUTED" {
		status = "UNKNOWN"
	}
	values := []struct{ id, expected, observed string }{
		{"receipt-digest", "valid", boolText(model.ValidDigest(receipt.Digest))},
		{"receipt-identity", model.ReceiptSchema, receipt.Schema},
		{"producer-identity", model.ProducerID, receipt.Producer},
		{"consumer-identity", model.ConsumerID, receipt.Consumer},
		{"meta-operation", model.MetaOperationID, receipt.MetaOperation},
		{"proof-choice", "non-empty", receipt.ProofChoice},
		{"state-determined", "fresh/stale/unknown", verdict.State},
		{"stage-attribution", "non-empty", verdict.Coordinate.Stage},
		{"temporal-boundary", "explicit", boolText(receipt.Boundary.ObservationEpoch > 0 && receipt.Boundary.ValidThroughEpoch >= receipt.Boundary.ObservationEpoch)},
		{"read-only-effects", "writes=0 authority=false", fmt.Sprintf("writes=%d authority=%t", receipt.RepositoryWrites, receipt.MutationAuthority)},
	}
	result := make([]model.CheckResult, len(values))
	for index, value := range values {
		checkStatus := status
		if status == "PASS" && value.observed == "" {
			checkStatus = "FAIL"
		}
		result[index] = model.CheckResult{ID: value.id, Status: checkStatus, Expected: value.expected,
			Observed: value.observed, Producer: model.ProducerID, Consumer: model.ConsumerID,
			MetaOperation: model.MetaOperationID, ProofChoice: receipt.ProofChoice}
	}
	return result
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
