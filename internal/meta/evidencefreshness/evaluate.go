package evidencefreshness

import (
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/compiler"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/decider"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/producer"
)

type Input struct {
	Contract     model.Contract
	HeadSHA      string
	Source       []byte
	Independence model.IndependenceEvidence
	WriteSet     model.WriteSetObservation
}

type observationCase struct {
	id              string
	class           string
	mutation        string
	source          []byte
	sourceAvailable bool
}

func Evaluate(input Input) model.Report {
	report := model.Report{
		Schema: model.ReportSchema, Scope: model.Scope, HeadSHA: input.HeadSHA,
		ContractDigest: model.DigestJSON(input.Contract),
		Independence:   input.Independence, IndependenceDigest: model.DigestJSON(input.Independence),
		NotClaimed: append([]string{}, input.Contract.NotClaimed...), WriteSet: input.WriteSet,
	}
	compiled, compileErr := compiler.Compile(input.Contract.SourcePath, input.Source)
	if compileErr == nil {
		report.Policy = compiled.Policy
		report.PolicyDigest = compiled.PolicyDigest
		report.SourceDigest = model.DigestBytes(input.Source)
		report.SemanticDigest = compiled.SemanticDigest
	}
	receipt, receiptErr := producer.BuildReceipt(input.Source, input.HeadSHA, input.Contract.BaseContext, input.Independence, input.WriteSet)
	if receiptErr != nil {
		compileErr = receiptErr
	} else {
		report.Receipt = receipt
		report.ReceiptDigest = receipt.Digest
	}
	if receiptErr == nil {
		report.Cases = evaluateCases(input, receipt)
		report.ClaimLedger = buildClaimLedger(report.Cases)
		report.ClaimLedgerDigest = model.DigestJSON(report.ClaimLedger)
	}
	report.Summary = summarize(report.Cases, input.Independence, input.WriteSet)
	report.Indicators = indicators(report, input.Contract)
	report.Decision, report.Resolution, report.Reason = model.DecisionFailClosed, model.ResolutionInvariant, "EVIDENCE_FRESHNESS_CONTRACT_MISMATCH"
	if compileErr == nil && reflect.DeepEqual(input.Contract, CanonicalContract()) && model.ValidHead(input.HeadSHA) &&
		len(input.Source) > 0 && receiptErr == nil && len(report.Cases) == model.CaseTotal &&
		len(report.ClaimLedger) == model.TransitionTotal && allIndicatorsSatisfied(report.Indicators) {
		report.Decision, report.Resolution, report.Reason = model.DecisionPass, model.ResolutionExact, "EVIDENCE_FRESHNESS_CONTRACT_SATISFIED"
	}
	report.Digest = reportDigest(report)
	return report
}

func evaluateCases(input Input, receipt model.Receipt) []model.CaseResult {
	cases := observationCases(input.Source)
	receiptRaw, err := model.Marshal(receipt)
	if err != nil {
		return nil
	}
	results := make([]model.CaseResult, 0, len(cases))
	for _, item := range cases {
		variantReceipt := receipt
		if item.sourceAvailable {
			if observed, observeErr := producer.BuildReceipt(item.source, input.HeadSHA, input.Contract.BaseContext, input.Independence, input.WriteSet); observeErr == nil {
				variantReceipt = observed
			}
		}
		context := contextForCase(input.Contract.BaseContext, receipt, variantReceipt, item.mutation)
		contextRaw, contextErr := model.Marshal(context)
		verdict := model.Verdict{}
		if contextErr != nil {
			verdict = model.Verdict{Schema: model.VerdictSchema, State: model.StateUnknown,
				Decision: model.DecisionFailClosed, Resolution: model.ResolutionLower,
				Reason: "CONTEXT_UNKNOWN", Coordinate: model.Coordinate{Stage: model.StageVerifier, Step: "marshal-context", Reason: "CONTEXT_UNKNOWN"}}
		} else {
			verdict = decider.Decide(item.source, receiptRaw, contextRaw)
		}
		results = append(results, caseResult(item, context, verdict))
	}
	return results
}

func observationCases(source []byte) []observationCase {
	commentSource := append([]byte("// CI intervention: presentation-only comment\n"), source...)
	semanticSource := []byte(strings.Replace(string(source),
		`entity ClaimSubject id "gooo://evidence-freshness/entity/claim-subject"`,
		`entity ClaimSubject id "gooo://evidence-freshness/entity/claim-subject-semantic-change"`, 1))
	return []observationCase{
		{id: "current-exact-head", class: model.ObservationCurrent, mutation: "", source: source, sourceAvailable: true},
		{id: "synthetic-comment-only", class: model.ObservationSynthetic, mutation: "comment-only", source: commentSource, sourceAvailable: true},
		{id: "synthetic-semantic-change", class: model.ObservationSynthetic, mutation: "semantic-change", source: semanticSource, sourceAvailable: true},
		{id: "synthetic-subject-change", class: model.ObservationSynthetic, mutation: "subject", source: source, sourceAvailable: true},
		{id: "synthetic-recipe-change", class: model.ObservationSynthetic, mutation: "recipe", source: source, sourceAvailable: true},
		{id: "synthetic-environment-change", class: model.ObservationSynthetic, mutation: "environment", source: source, sourceAvailable: true},
		{id: "synthetic-runner-change", class: model.ObservationSynthetic, mutation: "runner", source: source, sourceAvailable: true},
		{id: "synthetic-verifier-change", class: model.ObservationSynthetic, mutation: "verifier", source: source, sourceAvailable: true},
		{id: "synthetic-expired-boundary", class: model.ObservationSynthetic, mutation: "temporal-expired", source: source, sourceAvailable: true},
		{id: "synthetic-source-unavailable", class: model.ObservationSynthetic, mutation: "source-unavailable", source: nil, sourceAvailable: false},
	}
}

func contextForCase(base model.Context, receipt, variant model.Receipt, mutation string) model.Context {
	context := base
	context.Schema = model.ContextSchema
	context.PolicyDigest = variant.PolicyDigest
	context.Tuple = variant.Tuple
	context.EnvironmentBoundary = variant.Boundary.EnvironmentBoundary
	context.Consumer = model.ConsumerID
	switch mutation {
	case "subject":
		context.Tuple.Subject += "|changed"
	case "recipe":
		context.Tuple.Recipe += "|changed"
	case "environment":
		context.Tuple.Environment += "|changed"
		context.EnvironmentBoundary += "|changed"
	case "runner":
		context.Tuple.Runner += "|changed"
	case "verifier":
		context.Tuple.Verifier += "|changed"
	case "temporal-expired":
		context.CurrentEpoch = receipt.Boundary.ValidThroughEpoch + 1
	case "source-unavailable":
		context = contextForReceipt(base, receipt)
	}
	return context
}

func contextForReceipt(base model.Context, receipt model.Receipt) model.Context {
	context := base
	context.Schema = model.ContextSchema
	context.PolicyDigest = receipt.PolicyDigest
	context.Tuple = receipt.Tuple
	context.EnvironmentBoundary = receipt.Boundary.EnvironmentBoundary
	context.Consumer = model.ConsumerID
	return context
}

func caseResult(item observationCase, context model.Context, verdict model.Verdict) model.CaseResult {
	return model.CaseResult{ID: item.id, ObservationClass: item.class, Mutation: item.mutation, Status: "OBSERVED",
		ObservedState: verdict.State, ObservedDecision: verdict.Decision, ObservedResolution: verdict.Resolution,
		ObservedReason: verdict.Reason, RawFreshness: verdict.RawFreshness, SemanticFreshness: verdict.SemanticFreshness,
		SourceAvailable: item.sourceAvailable, SourceDigest: verdict.SourceDigest, SemanticDigest: verdict.SemanticDigest,
		Coordinate: verdict.Coordinate, Context: context, ChangedDimensions: verdict.ChangedDimensions,
		Transition: verdict.Transition, Checks: verdict.Checks}
}

func buildClaimLedger(cases []model.CaseResult) []model.ClaimLedgerEntry {
	entries := make([]model.ClaimLedgerEntry, 0, len(cases))
	previous := ""
	for index, item := range cases {
		nextState, preservation := model.ClaimOpen, "PRESERVE_OPEN_UNSUPPORTED_CURRENT"
		switch item.ObservedState {
		case model.StateFresh:
			nextState, preservation = model.ClaimDischarged, "DISCHARGE_CURRENT_EVIDENCE"
		case model.StateUnknown:
			preservation = "PRESERVE_OPEN_UNKNOWN"
		case model.StateRefuted:
			nextState, preservation = model.ClaimRefuted, "EXPLICIT_REFUTATION"
		}
		entry := model.ClaimLedgerEntry{Schema: model.LedgerSchema, Sequence: index + 1,
			ClaimID: item.Transition.ClaimID, PriorState: model.ClaimOpen, NextState: nextState,
			Preservation: preservation, FreshnessObservation: item.ObservedState,
			ReceiptDigest: item.Transition.EvidenceDigest, SourceDigest: item.SourceDigest,
			SemanticDigest: item.SemanticDigest, PreviousDigest: previous,
			Provenance: []string{item.Transition.EvidenceDigest, item.SourceDigest, item.SemanticDigest, item.Coordinate.Stage, item.Coordinate.Step, item.ObservedReason}}
		entry = model.SealLedgerEntry(entry)
		entries = append(entries, entry)
		previous = entry.Digest
	}
	return entries
}

func summarize(cases []model.CaseResult, independence model.IndependenceEvidence, writeSet model.WriteSetObservation) model.Summary {
	summary := model.Summary{CasesObserved: len(cases), FixedAxisDenominator: model.AxisTotal,
		RawStaleByStage: map[string]int{}, StaleByStage: map[string]int{}, UnknownByStage: map[string]int{},
		ForbiddenDependencyCount: independence.ForbiddenDependencyCount,
		IndependenceContract:     independence.IndependenceContract,
		ReadOnlyBeforeCount:      writeSet.BeforeCount, ReadOnlyAfterCount: writeSet.AfterCount,
		ReadOnlyWriteSetStable: writeSet.BeforeDigest != "" && writeSet.BeforeDigest == writeSet.AfterDigest && writeSet.BeforeCount == writeSet.AfterCount}
	axis := map[string]bool{}
	for _, item := range cases {
		if item.ObservationClass == model.ObservationCurrent {
			summary.CurrentEvidenceCases++
		} else if item.ObservationClass == model.ObservationSynthetic {
			summary.SyntheticCounterexamples++
		}
		countState(item.RawFreshness, &summary.RawFreshCases, &summary.RawStaleCases, &summary.RawUnknownCases)
		countState(item.SemanticFreshness, &summary.SemanticFreshCases, &summary.SemanticStaleCases, &summary.SemanticUnknownCases)
		switch item.ObservedState {
		case model.StateFresh:
			summary.ClaimFreshCases++
			summary.ClaimDischarged++
		case model.StateStale:
			summary.ClaimStaleCases++
			summary.ClaimOpenPreserved++
			summary.StaleByStage[item.Coordinate.Stage]++
		case model.StateUnknown:
			summary.ClaimUnknownCases++
			summary.ClaimOpenPreserved++
			summary.UnknownByStage[item.Coordinate.Stage]++
		case model.StateRefuted:
			summary.ClaimRefuted++
		}
		if item.RawFreshness == model.StateStale {
			summary.RawStaleByStage[item.Coordinate.Stage]++
		}
		for _, dimension := range item.ChangedDimensions {
			for _, axisName := range []string{"subject", "material", "recipe", "environment", "runner", "verifier"} {
				if dimension == axisName || strings.HasPrefix(dimension, axisName+".") {
					axis[axisName] = true
				}
			}
		}
		if item.Transition.From != "" {
			summary.FreshnessTransitions++
		}
		summary.ClaimLedgerEntries++
		if item.SourceAvailable && model.ValidDigest(item.SourceDigest) && model.ValidDigest(item.SemanticDigest) {
			summary.SourceReconstructedCases++
		}
		if !item.SourceAvailable {
			summary.SourceUnavailableCases++
		}
	}
	summary.AxisChangesObserved = len(axis)
	return summary
}

func countState(state string, fresh, stale, unknown *int) {
	switch state {
	case model.StateFresh:
		*fresh = *fresh + 1
	case model.StateStale:
		*stale = *stale + 1
	case model.StateUnknown:
		*unknown = *unknown + 1
	}
}

func indicators(report model.Report, contract model.Contract) []model.Indicator {
	values := map[string]int{
		"gooo.metric.evidence-freshness.cases.v2":                          report.Summary.CasesObserved,
		"gooo.metric.evidence-freshness.current-evidence.v2":               report.Summary.CurrentEvidenceCases,
		"gooo.metric.evidence-freshness.synthetic-counterexamples.v2":      report.Summary.SyntheticCounterexamples,
		"gooo.metric.evidence-freshness.coupling-axes.v2":                  report.Summary.AxisChangesObserved,
		"gooo.metric.evidence-freshness.raw-source-reconstruction.v2":      report.Summary.SourceReconstructedCases,
		"gooo.metric.evidence-freshness.semantic-source-reconstruction.v2": report.Summary.SourceReconstructedCases,
		"gooo.metric.evidence-freshness.comment-intervention.v2":           boolInt(commentInterventionPassed(report.Cases)),
		"gooo.metric.evidence-freshness.semantic-intervention.v2":          boolInt(semanticInterventionPassed(report.Cases)),
		"gooo.metric.evidence-freshness.freshness-transitions.v2":          report.Summary.FreshnessTransitions,
		"gooo.metric.evidence-freshness.claim-ledger.v2":                   len(report.ClaimLedger),
		"gooo.metric.evidence-freshness.source-unavailable.v2":             boolInt(sourceUnavailablePassed(report.Cases)),
		"gooo.metric.evidence-freshness.read-only-before-after.v2":         boolInt(report.Summary.ReadOnlyWriteSetStable && report.Summary.ReadOnlyBeforeCount == 0 && report.Summary.ReadOnlyAfterCount == 0),
		"gooo.metric.evidence-freshness.independence-contract.v2":          boolInt(independenceContractSatisfied(report.Independence)),
	}
	result := make([]model.Indicator, len(contract.Metrics))
	for index, definition := range contract.Metrics {
		numerator := values[definition.MetricID]
		basisPoints := 0
		if definition.Denominator > 0 {
			basisPoints = numerator * 10000 / definition.Denominator
		}
		result[index] = model.Indicator{MetricID: definition.MetricID, Class: definition.Class,
			Producer: definition.Producer, Consumer: definition.Consumer, ProofChoice: definition.ProofChoice,
			MetaOperation: definition.MetaOperation, Numerator: numerator, Denominator: definition.Denominator,
			BasisPoints: basisPoints, ExpectedNumerator: definition.ExpectedNumerator,
			Satisfied: numerator == definition.ExpectedNumerator}
	}
	return result
}

func commentInterventionPassed(cases []model.CaseResult) bool {
	item, ok := findCase(cases, "synthetic-comment-only")
	return ok && item.RawFreshness == model.StateStale && item.SemanticFreshness == model.StateFresh &&
		item.ObservedState == model.StateFresh && item.ObservedDecision == model.DecisionPass && item.Transition.To == "CLAIM_PRESERVED"
}

func semanticInterventionPassed(cases []model.CaseResult) bool {
	item, ok := findCase(cases, "synthetic-semantic-change")
	return ok && item.RawFreshness == model.StateStale && item.SemanticFreshness == model.StateStale &&
		item.ObservedState == model.StateStale && item.ObservedDecision == model.DecisionFailClosed
}

func sourceUnavailablePassed(cases []model.CaseResult) bool {
	item, ok := findCase(cases, "synthetic-source-unavailable")
	return ok && !item.SourceAvailable && item.ObservedState == model.StateUnknown &&
		item.ObservedDecision == model.DecisionFailClosed && item.ObservedResolution == model.ResolutionLower
}

func findCase(cases []model.CaseResult, id string) (model.CaseResult, bool) {
	for _, item := range cases {
		if item.ID == id {
			return item, true
		}
	}
	return model.CaseResult{}, false
}

func allIndicatorsSatisfied(indicators []model.Indicator) bool {
	if len(indicators) != model.MetricTotal {
		return false
	}
	for _, item := range indicators {
		if !item.Satisfied || item.Denominator <= 0 || item.BasisPoints != 10000 {
			return false
		}
	}
	return true
}

func independenceContractSatisfied(evidence model.IndependenceEvidence) bool {
	return evidence.Schema == model.IndependenceSchema && evidence.ForbiddenDependencyCount == 0 &&
		evidence.IndependenceContract.Numerator == model.IndependenceContractTotal &&
		evidence.IndependenceContract.Denominator == model.IndependenceContractTotal
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func reportDigest(report model.Report) string {
	report.Digest = ""
	return model.DigestJSON(report)
}
