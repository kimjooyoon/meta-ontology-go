package evidencefreshness

import (
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
)

func Validate(report model.Report) error {
	contract := CanonicalContract()
	if report.Schema != model.ReportSchema || report.Scope != model.Scope || !model.ValidHead(report.HeadSHA) ||
		report.ContractDigest != model.DigestJSON(contract) || !model.ValidDigest(report.SourceDigest) ||
		report.Receipt.Schema != model.ReceiptSchema || !model.VerifyReceiptDigest(report.Receipt) ||
		report.ReceiptDigest != report.Receipt.Digest || report.Receipt.HeadSHA != report.HeadSHA ||
		!reflect.DeepEqual(report.Receipt.Independence, report.Independence) ||
		!reflect.DeepEqual(report.Independence, model.DefaultIndependenceEvidence()) ||
		report.IndependenceDigest != model.DigestJSON(report.Independence) {
		return fmt.Errorf("EVIDENCE_FRESHNESS_IDENTITY_MISMATCH")
	}
	if report.Decision != model.DecisionPass || report.Resolution != model.ResolutionExact ||
		report.Reason != "EVIDENCE_FRESHNESS_CONTRACT_SATISFIED" || len(report.Cases) != model.CaseTotal ||
		len(report.Indicators) != model.MetricTotal || report.RepositoryWrites != 0 || report.MutationAuthority {
		return fmt.Errorf("EVIDENCE_FRESHNESS_DECISION_OR_DENOMINATOR_MISMATCH")
	}
	wantSummary := model.Summary{CasesSatisfied: model.CaseTotal, CasesTotal: model.CaseTotal,
		FreshCases: 1, StaleCases: 7, UnknownCases: 2, AxisChangesObserved: model.AxisTotal,
		FixedAxisDenominator: model.AxisTotal, StaleByStage: map[string]int{
			model.StageSubject: 1, model.StageMaterial: 1, model.StageRecipe: 1,
			model.StageEnvironment: 1, model.StageRunner: 1, model.StageVerifier: 2,
		}, UnknownByStage: map[string]int{model.StageSubject: 1, model.StageVerifier: 1},
		PreservationTransitions: model.TransitionTotal, TemporalBoundaryCases: 1, ReadOnlyCases: 1,
		ForbiddenDependencyCount: 0,
		IndependenceContract:     model.FixedMetric{Numerator: model.IndependenceContractTotal, Denominator: model.IndependenceContractTotal}}
	if !reflect.DeepEqual(report.Summary, wantSummary) {
		return fmt.Errorf("EVIDENCE_FRESHNESS_SUMMARY_MISMATCH")
	}
	for index, item := range report.Cases {
		definition := contract.Cases[index]
		if item.Status != "SATISFIED" || item.ID != definition.ID || item.Mutation != definition.Mutation ||
			item.ExpectedState != definition.ExpectedState || item.ExpectedDecision != definition.ExpectedDecision ||
			item.ExpectedResolution != definition.ExpectedResolution || item.ExpectedStage != definition.ExpectedStage ||
			item.ExpectedStep != definition.ExpectedStep || item.ExpectedReason != definition.ExpectedReason ||
			item.ObservedState != definition.ExpectedState || item.ObservedDecision != definition.ExpectedDecision ||
			item.ObservedResolution != definition.ExpectedResolution || item.ObservedReason != definition.ExpectedReason ||
			len(item.Checks) != model.CheckTotal || item.Transition.From != "CLAIM_JUSTIFIED" || item.Transition.Coordinate.Stage == "" {
			return fmt.Errorf("EVIDENCE_FRESHNESS_CASE_MISMATCH_%d", index)
		}
		if item.Transition.Coordinate != item.Coordinate || item.Transition.EvidenceDigest != report.ReceiptDigest {
			return fmt.Errorf("EVIDENCE_FRESHNESS_TRANSITION_MISMATCH_%d", index)
		}
		if item.ObservedState == model.StateFresh && (item.Transition.To != "CLAIM_PRESERVED" || item.Transition.Preservation != "PRESERVE_EXACT") {
			return fmt.Errorf("EVIDENCE_FRESHNESS_FRESH_TRANSITION_MISMATCH_%d", index)
		}
		if item.ObservedState == model.StateStale && (item.Transition.To != "CLAIM_STALE" || item.Transition.Preservation != "DO_NOT_PRESERVE") {
			return fmt.Errorf("EVIDENCE_FRESHNESS_STALE_TRANSITION_MISMATCH_%d", index)
		}
		if item.ObservedState == model.StateUnknown && (item.Transition.To != "CLAIM_UNKNOWN" || item.Transition.Preservation != "DO_NOT_PRESERVE") {
			return fmt.Errorf("EVIDENCE_FRESHNESS_UNKNOWN_TRANSITION_MISMATCH_%d", index)
		}
	}
	for index, definition := range contract.Metrics {
		item := report.Indicators[index]
		if item.MetricID != definition.MetricID || item.Class != definition.Class || item.Producer != definition.Producer ||
			item.Consumer != definition.Consumer || item.ProofChoice != definition.ProofChoice || item.MetaOperation != definition.MetaOperation ||
			item.Numerator != definition.ExpectedNumerator || item.Denominator != definition.Denominator || item.ExpectedNumerator != definition.ExpectedNumerator || !item.Satisfied ||
			item.BasisPoints != 10000 {
			return fmt.Errorf("EVIDENCE_FRESHNESS_INDICATOR_MISMATCH_%d", index)
		}
	}
	if !reflect.DeepEqual(report.NotClaimed, contract.NotClaimed) || report.Digest != reportDigest(report) {
		return fmt.Errorf("EVIDENCE_FRESHNESS_DIGEST_OR_CLAIM_MISMATCH")
	}
	return nil
}
