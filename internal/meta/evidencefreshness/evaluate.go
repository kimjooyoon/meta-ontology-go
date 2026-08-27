package evidencefreshness

import (
	"reflect"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/decider"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/producer"
)

type Input struct {
	Contract     model.Contract
	HeadSHA      string
	Source       []byte
	Independence model.IndependenceEvidence
}

func Evaluate(input Input) model.Report {
	report := model.Report{Schema: model.ReportSchema, Scope: model.Scope, HeadSHA: input.HeadSHA,
		ContractDigest: model.DigestJSON(input.Contract), SourceDigest: model.DigestBytes(input.Source),
		Independence: input.Independence, IndependenceDigest: model.DigestJSON(input.Independence), NotClaimed: append([]string{}, input.Contract.NotClaimed...),
		RepositoryWrites: 0, MutationAuthority: false}
	receipt, err := producer.BuildReceipt(input.Source, input.HeadSHA, input.Contract.BaseContext)
	if err == nil {
		report.Receipt = receipt
		report.ReceiptDigest = receipt.Digest
		receiptRaw, marshalErr := model.Marshal(receipt)
		if marshalErr == nil {
			for _, definition := range input.Contract.Cases {
				context := ContextForMutation(input.Contract.BaseContext, receipt, definition.Mutation)
				contextRaw, contextErr := model.Marshal(context)
				verdict := model.Verdict{}
				if contextErr != nil {
					verdict = model.Verdict{Schema: model.VerdictSchema, State: model.StateUnknown,
						Decision: model.DecisionFailClosed, Resolution: model.ResolutionLower,
						Reason: "CONTEXT_UNKNOWN"}
				} else {
					verdict = decider.Decide(receiptRaw, contextRaw)
				}
				report.Cases = append(report.Cases, caseResult(definition, context, verdict))
			}
		}
	}
	report.Summary = summarize(report.Cases, input.Independence)
	report.Indicators = indicators(report, input.Contract)
	report.Decision, report.Resolution, report.Reason = model.DecisionFailClosed, model.ResolutionInvariant, "EVIDENCE_FRESHNESS_CONTRACT_MISMATCH"
	if reflect.DeepEqual(input.Contract, CanonicalContract()) && model.ValidHead(input.HeadSHA) && len(input.Source) > 0 &&
		err == nil && len(report.Cases) == model.CaseTotal && input.Independence.Schema == model.IndependenceSchema &&
		input.Independence.ProducerDependencies == 0 && input.Independence.DeciderDependencies == 0 &&
		report.Summary.CasesSatisfied == model.CaseTotal {
		report.Decision, report.Resolution, report.Reason = model.DecisionPass, model.ResolutionExact, "EVIDENCE_FRESHNESS_CONTRACT_SATISFIED"
	}
	report.Digest = reportDigest(report)
	return report
}

func ContextForMutation(base model.Context, receipt model.Receipt, mutation string) model.Context {
	context := base
	context.Schema = model.ContextSchema
	context.Tuple = receipt.Tuple
	context.Consumer = model.ConsumerID
	switch mutation {
	case "subject":
		context.Tuple.Subject += "|changed"
	case "material":
		context.Tuple.Material += "|changed"
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
	case "unknown-subject":
		context.Tuple.Subject = ""
	case "unknown-verifier":
		context.Tuple.Verifier = ""
	}
	return context
}

func caseResult(definition model.CaseDefinition, context model.Context, verdict model.Verdict) model.CaseResult {
	status := "NOT_SATISFIED"
	if verdict.State == definition.ExpectedState && verdict.Decision == definition.ExpectedDecision &&
		verdict.Resolution == definition.ExpectedResolution && verdict.Coordinate.Stage == definition.ExpectedStage &&
		verdict.Coordinate.Step == definition.ExpectedStep && verdict.Reason == definition.ExpectedReason {
		status = "SATISFIED"
	}
	return model.CaseResult{ID: definition.ID, Status: status, Mutation: definition.Mutation,
		ExpectedState: definition.ExpectedState, ExpectedDecision: definition.ExpectedDecision,
		ExpectedResolution: definition.ExpectedResolution, ExpectedStage: definition.ExpectedStage,
		ExpectedStep: definition.ExpectedStep, ExpectedReason: definition.ExpectedReason,
		ObservedState: verdict.State, ObservedDecision: verdict.Decision,
		ObservedResolution: verdict.Resolution, ObservedReason: verdict.Reason,
		Coordinate: verdict.Coordinate, Context: context, ChangedDimensions: verdict.ChangedDimensions,
		Transition: verdict.Transition, Checks: verdict.Checks}
}

func summarize(cases []model.CaseResult, independence model.IndependenceEvidence) model.Summary {
	summary := model.Summary{CasesTotal: model.CaseTotal, FixedAxisDenominator: model.AxisTotal,
		StaleByStage: map[string]int{}, UnknownByStage: map[string]int{}, ReadOnlyCases: 1}
	axis := map[string]bool{}
	for _, item := range cases {
		if item.Status == "SATISFIED" {
			summary.CasesSatisfied++
		}
		switch item.ObservedState {
		case model.StateFresh:
			summary.FreshCases++
		case model.StateStale:
			summary.StaleCases++
			summary.StaleByStage[item.Coordinate.Stage]++
		case model.StateUnknown:
			summary.UnknownCases++
			summary.UnknownByStage[item.Coordinate.Stage]++
		}
		for _, dimension := range item.ChangedDimensions {
			if dimension == "subject" || dimension == "material" || dimension == "recipe" || dimension == "environment" || dimension == "runner" || dimension == "verifier" {
				axis[dimension] = true
			}
		}
		if item.Transition.From != "" {
			summary.PreservationTransitions++
		}
		if item.Mutation == "temporal-expired" {
			summary.TemporalBoundaryCases++
		}
	}
	summary.AxisChangesObserved = len(axis)
	_ = independence
	return summary
}

func indicators(report model.Report, contract model.Contract) []model.Indicator {
	values := make(map[string]int, len(contract.Metrics))
	values["gooo.metric.evidence-freshness.cases.v1"] = report.Summary.CasesSatisfied
	values["gooo.metric.evidence-freshness.fresh.v1"] = report.Summary.FreshCases
	values["gooo.metric.evidence-freshness.stale.v1"] = report.Summary.StaleCases
	values["gooo.metric.evidence-freshness.unknown.v1"] = report.Summary.UnknownCases
	values["gooo.metric.evidence-freshness.coupling-axes.v1"] = report.Summary.AxisChangesObserved
	values["gooo.metric.evidence-freshness.stage-attribution.v1"] = distinctStages(report.Summary.StaleByStage)
	values["gooo.metric.evidence-freshness.transitions.v1"] = report.Summary.PreservationTransitions
	values["gooo.metric.evidence-freshness.temporal-boundary.v1"] = report.Summary.TemporalBoundaryCases
	values["gooo.metric.evidence-freshness.read-only.v1"] = boolInt(report.RepositoryWrites == 0 && !report.MutationAuthority)
	values["gooo.metric.evidence-freshness.independent-decider.v1"] = 0
	if report.Independence.Schema == model.IndependenceSchema && report.Independence.ProducerDependencies == 0 && report.Independence.DeciderDependencies == 0 {
		values["gooo.metric.evidence-freshness.independent-decider.v1"] = 1
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

func distinctStages(counts map[string]int) int {
	stages := make([]string, 0, len(counts))
	for stage := range counts {
		if counts[stage] > 0 {
			stages = append(stages, stage)
		}
	}
	sort.Strings(stages)
	return len(stages)
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
