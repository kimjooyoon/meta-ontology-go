package decider

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/compiler"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
)

// Decide independently reconstructs the supplied source through the
// canonical parser/lowerer. It intentionally does not import producer or the
// evaluator; the receipt is evidence to compare, not decision authority.
func Decide(sourceRaw, receiptRaw, contextRaw []byte) model.Verdict {
	receipt, receiptErr := model.DecodeStrict[model.Receipt](receiptRaw)
	context, contextErr := model.DecodeStrict[model.Context](contextRaw)
	if receiptErr != nil || contextErr != nil {
		return finish(unknown("SOURCE_OR_CONTEXT_UNKNOWN", model.StageSubject, "decode-receipt-or-context", "source-or-context"), receipt, context)
	}
	if receipt.Schema != model.ReceiptSchema || !model.ValidHead(receipt.HeadSHA) || !model.VerifyReceiptDigest(receipt) {
		return finish(unknown("RECEIPT_UNKNOWN", model.StageSubject, "verify-receipt", "receipt-identity"), receipt, context)
	}
	if context.Schema != model.ContextSchema || context.Consumer != model.ConsumerID {
		return finish(unknown("CONSUMER_UNKNOWN", model.StageVerifier, "verify-context", "consumer-identity"), receipt, context)
	}
	if receipt.Producer != model.ProducerID || receipt.Consumer != model.ConsumerID || receipt.MetaOperation != model.MetaOperationID ||
		receipt.ProofChoice == "" || receipt.PriorClaimState != model.ClaimOpen {
		return finish(refuted("RECEIPT_META_BINDING_INVALID", model.StageVerifier, "verify-receipt-meta", "meta-binding"), receipt, context)
	}
	if !readOnly(receipt.WriteSet) {
		return finish(refuted("READ_ONLY_EFFECT_INVALID", model.StageVerifier, "verify-before-after-write-set", "read-only-effects"), receipt, context)
	}
	if sourceRaw == nil {
		return finish(unknown("SOURCE_UNAVAILABLE", model.StageSubject, "reconstruct-source", "source"), receipt, context)
	}
	compiled, err := compiler.Compile(receipt.SourcePath, sourceRaw)
	if err != nil {
		return finish(unknown("SOURCE_RECONSTRUCTION_UNKNOWN", model.StageSubject, "parse-and-lower-source", "source"), receipt, context)
	}
	if compiled.PolicyDigest != receipt.PolicyDigest || context.PolicyDigest != receipt.PolicyDigest {
		return finish(unknown("POLICY_UNKNOWN", model.StageVerifier, "compare-policy-digest", "policy"), receipt, context, compiled.SemanticDigest, model.DigestBytes(sourceRaw))
	}
	rawDigest := model.DigestBytes(sourceRaw)
	semanticDigest := compiled.SemanticDigest
	if context.Tuple.Material.RawDigest != rawDigest || context.Tuple.Material.SemanticDigest != semanticDigest {
		return finish(withDigests(unknown("CURRENT_CONTEXT_MATERIAL_MISMATCH", model.StageMaterial, "compare-current-material", "current-material"), semanticDigest, rawDigest), receipt, context)
	}
	rawFreshness := model.StateFresh
	semanticFreshness := model.StateFresh
	changed := []string{}
	if rawDigest != receipt.Tuple.Material.RawDigest {
		rawFreshness = model.StateStale
		changed = append(changed, "material.raw")
	}
	if semanticDigest != receipt.Tuple.Material.SemanticDigest {
		semanticFreshness = model.StateStale
		changed = append(changed, "material.semantic")
	}
	if rawFreshness == model.StateStale && semanticFreshness == model.StateFresh && compiled.Policy.SemanticPolicy == "comments_ignored" {
		return finish(withDigests(model.Verdict{Schema: model.VerdictSchema, State: model.StateFresh,
			Decision: model.DecisionPass, Resolution: model.ResolutionExact,
			Reason:            "RAW_MATERIAL_CHANGED_SEMANTIC_PRESERVED",
			Coordinate:        model.Coordinate{Stage: model.StageMaterial, Step: "compare-raw-material", Reason: "RAW_MATERIAL_CHANGED_SEMANTIC_PRESERVED"},
			ChangedDimensions: changed, RawFreshness: rawFreshness, SemanticFreshness: semanticFreshness}, semanticDigest, rawDigest), receipt, context)
	}
	if semanticFreshness == model.StateStale {
		return finish(withDigests(stale("SEMANTIC_DIGEST_CHANGED", model.StageMaterial, "compare-semantic-material", changed, rawFreshness, semanticFreshness), semanticDigest, rawDigest), receipt, context)
	}
	if rawFreshness == model.StateStale {
		return finish(withDigests(stale("RAW_MATERIAL_CHANGED", model.StageMaterial, "compare-raw-material", changed, rawFreshness, semanticFreshness), semanticDigest, rawDigest), receipt, context)
	}
	for _, axis := range compiled.Policy.Axes {
		receiptValue, receiptKnown := axisValue(receipt.Tuple, axis.Name)
		currentValue, currentKnown := axisValue(context.Tuple, axis.Name)
		if !receiptKnown || !currentKnown {
			return finish(withDigests(unknown(strings.ToUpper(axis.Name)+"_UNKNOWN", axis.Stage, "read-"+axis.Name, axis.Name), semanticDigest, rawDigest), receipt, context)
		}
		if receiptValue != currentValue {
			changed = append(changed, axis.Name)
			return finish(withDigests(stale(strings.ToUpper(axis.Name)+"_CHANGED", axis.Stage, axis.Step, changed, rawFreshness, semanticFreshness), semanticDigest, rawDigest), receipt, context)
		}
	}
	if receipt.Boundary.EnvironmentBoundary == "" || context.EnvironmentBoundary == "" {
		return finish(withDigests(unknown("ENVIRONMENT_BOUNDARY_UNKNOWN", model.StageEnvironment, "read-environment-boundary", "environment-boundary"), semanticDigest, rawDigest), receipt, context)
	}
	if receipt.Boundary.EnvironmentBoundary != context.EnvironmentBoundary {
		return finish(withDigests(stale("ENVIRONMENT_BOUNDARY_CHANGED", model.StageEnvironment, "compare-environment-boundary", []string{"environment-boundary"}, rawFreshness, semanticFreshness), semanticDigest, rawDigest), receipt, context)
	}
	if receipt.Boundary.ObservationEpoch <= 0 || receipt.Boundary.ValidThroughEpoch < receipt.Boundary.ObservationEpoch || context.CurrentEpoch <= 0 {
		return finish(withDigests(unknown("TEMPORAL_BOUNDARY_UNKNOWN", model.StageVerifier, "read-validity-boundary", "temporal-boundary"), semanticDigest, rawDigest), receipt, context)
	}
	if context.CurrentEpoch < receipt.Boundary.ObservationEpoch {
		return finish(withDigests(unknown("TEMPORAL_BOUNDARY_PRECEDES_OBSERVATION", model.StageVerifier, "compare-validity-boundary", "temporal-boundary"), semanticDigest, rawDigest), receipt, context)
	}
	if context.CurrentEpoch > receipt.Boundary.ValidThroughEpoch {
		return finish(withDigests(stale("TEMPORAL_BOUNDARY_EXPIRED", model.StageVerifier, "check-validity-boundary", []string{"temporal-boundary"}, rawFreshness, semanticFreshness), semanticDigest, rawDigest), receipt, context)
	}
	return finish(withDigests(model.Verdict{Schema: model.VerdictSchema, State: model.StateFresh,
		Decision: model.DecisionPass, Resolution: model.ResolutionExact,
		Reason:       "TUPLE_EXACT_AND_BOUNDARY_CURRENT",
		Coordinate:   model.Coordinate{Stage: model.StageVerifier, Step: "accept-current-evidence", Reason: "TUPLE_EXACT_AND_BOUNDARY_CURRENT"},
		RawFreshness: rawFreshness, SemanticFreshness: semanticFreshness}, semanticDigest, rawDigest), receipt, context)
}

func axisValue(tuple model.EvidenceTuple, name string) (string, bool) {
	switch name {
	case "subject":
		return tuple.Subject, tuple.Subject != ""
	case "material":
		value := tuple.Material.RawDigest + "|" + tuple.Material.SemanticDigest
		return value, tuple.Material.RawDigest != "" && tuple.Material.SemanticDigest != ""
	case "recipe":
		return tuple.Recipe, tuple.Recipe != ""
	case "environment":
		return tuple.Environment, tuple.Environment != ""
	case "runner":
		return tuple.Runner, tuple.Runner != ""
	case "verifier":
		return tuple.Verifier, tuple.Verifier != ""
	default:
		return "", false
	}
}

func withDigests(verdict model.Verdict, semanticDigest, rawDigest string) model.Verdict {
	verdict.SemanticDigest, verdict.SourceDigest = semanticDigest, rawDigest
	return verdict
}

func stale(reason, stage, step string, dimensions []string, rawFreshness, semanticFreshness string) model.Verdict {
	return model.Verdict{Schema: model.VerdictSchema, State: model.StateStale,
		Decision: model.DecisionFailClosed, Resolution: model.ResolutionInvariant, Reason: reason,
		Coordinate: model.Coordinate{Stage: stage, Step: step, Reason: reason}, ChangedDimensions: dimensions,
		RawFreshness: rawFreshness, SemanticFreshness: semanticFreshness}
}

func unknown(reason, stage, step, dimension string) model.Verdict {
	return model.Verdict{Schema: model.VerdictSchema, State: model.StateUnknown,
		Decision: model.DecisionFailClosed, Resolution: model.ResolutionLower, Reason: reason,
		Coordinate: model.Coordinate{Stage: stage, Step: step, Reason: reason}, ChangedDimensions: []string{dimension},
		RawFreshness: model.StateUnknown, SemanticFreshness: model.StateUnknown}
}

func refuted(reason, stage, step, dimension string) model.Verdict {
	return model.Verdict{Schema: model.VerdictSchema, State: model.StateRefuted,
		Decision: model.DecisionFailClosed, Resolution: model.ResolutionInvariant, Reason: reason,
		Coordinate: model.Coordinate{Stage: stage, Step: step, Reason: reason}, ChangedDimensions: []string{dimension},
		RawFreshness: model.StateUnknown, SemanticFreshness: model.StateUnknown}
}

func finish(verdict model.Verdict, receipt model.Receipt, context model.Context, optional ...string) model.Verdict {
	if len(optional) >= 2 {
		verdict.SemanticDigest, verdict.SourceDigest = optional[0], optional[1]
	}
	verdict.ReceiptDigest = receipt.Digest
	verdict.ContextDigest = model.DigestJSON(context)
	verdict.Transition = transition(verdict, receipt.Digest)
	verdict.Checks = checks(verdict, receipt)
	return model.SealVerdict(verdict)
}

func transition(verdict model.Verdict, receiptDigest string) model.ClaimTransition {
	to, preservation := "CLAIM_UNKNOWN", "DO_NOT_PRESERVE"
	switch verdict.State {
	case model.StateFresh:
		to, preservation = "CLAIM_PRESERVED", "PRESERVE_EXACT"
	case model.StateStale:
		to = "CLAIM_STALE"
	case model.StateRefuted:
		to = "CLAIM_REFUTED"
	}
	return model.ClaimTransition{ClaimID: "gooo://evidence-freshness/claim/checked-source", From: "CLAIM_JUSTIFIED",
		To: to, Preservation: preservation, Coordinate: verdict.Coordinate, EvidenceDigest: receiptDigest}
}

func checks(verdict model.Verdict, receipt model.Receipt) []model.CheckResult {
	status := "PASS"
	if verdict.State == model.StateUnknown || verdict.State == model.StateRefuted || verdict.Decision != model.DecisionPass {
		status = "FAIL_CLOSED"
	}
	values := []struct{ id, expected, observed string }{
		{"source-reconstruction", "canonical parse/lower", verdict.SourceDigest},
		{"raw-material-digest", "material.raw_digest", receipt.Tuple.Material.RawDigest},
		{"semantic-digest", "material.semantic_digest", receipt.Tuple.Material.SemanticDigest},
		{"receipt-identity", model.ReceiptSchema, receipt.Schema},
		{"policy-identity", "policy digest", receipt.PolicyDigest},
		{"consumer-identity", model.ConsumerID, receipt.Consumer},
		{"prior-claim-state", model.ClaimOpen, receipt.PriorClaimState},
		{"boundary-policy", "explicit epoch/environment", fmt.Sprintf("%d..%d/%s", receipt.Boundary.ObservationEpoch, receipt.Boundary.ValidThroughEpoch, receipt.Boundary.EnvironmentBoundary)},
		{"read-only-before-after", "stable writes=0", fmt.Sprintf("%d->%d", receipt.WriteSet.BeforeCount, receipt.WriteSet.AfterCount)},
		{"state-determined", "fresh/stale/unknown", verdict.State},
	}
	result := make([]model.CheckResult, len(values))
	for index, value := range values {
		result[index] = model.CheckResult{ID: value.id, Status: status, Expected: value.expected, Observed: value.observed,
			Producer: model.ProducerID, Consumer: model.ConsumerID, MetaOperation: model.MetaOperationID, ProofChoice: receipt.ProofChoice}
	}
	return result
}

func readOnly(writeSet model.WriteSetObservation) bool {
	return model.ValidDigest(writeSet.BeforeDigest) && model.ValidDigest(writeSet.AfterDigest) &&
		writeSet.BeforeDigest == writeSet.AfterDigest && writeSet.BeforeCount == 0 && writeSet.AfterCount == 0
}
