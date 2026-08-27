package evidencefreshness

import (
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/compiler"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"
)

func Validate(report model.Report) error {
	contract := CanonicalContract()
	if compiler.ValidatePolicy(report.Policy) != nil || report.Schema != model.ReportSchema || report.Scope != model.Scope || !model.ValidHead(report.HeadSHA) ||
		report.ContractDigest != model.DigestJSON(contract) || report.Policy.Schema != model.PolicySchema ||
		report.PolicyDigest != model.DigestJSON(report.Policy) || !model.ValidDigest(report.SourceDigest) || !model.ValidDigest(report.SemanticDigest) ||
		report.Receipt.Schema != model.ReceiptSchema || !model.VerifyReceiptDigest(report.Receipt) ||
		report.ReceiptDigest != report.Receipt.Digest || report.Receipt.HeadSHA != report.HeadSHA ||
		report.Receipt.PolicyDigest != report.PolicyDigest || report.Receipt.SourceDigest != report.SourceDigest ||
		report.Receipt.SemanticDigest != report.SemanticDigest || !reflect.DeepEqual(report.Receipt.Independence, report.Independence) ||
		!reflect.DeepEqual(report.Independence, model.DefaultIndependenceEvidence()) ||
		!reflect.DeepEqual(report.Receipt.WriteSet, report.WriteSet) || report.IndependenceDigest != model.DigestJSON(report.Independence) {
		return fmt.Errorf("EVIDENCE_FRESHNESS_IDENTITY_MISMATCH")
	}
	if report.Decision != model.DecisionPass || report.Resolution != model.ResolutionExact ||
		report.Reason != "EVIDENCE_FRESHNESS_CONTRACT_SATISFIED" || len(report.Cases) != model.CaseTotal ||
		len(report.Indicators) != model.MetricTotal || len(report.ClaimLedger) != model.TransitionTotal ||
		!validWriteSet(report.WriteSet) {
		return fmt.Errorf("EVIDENCE_FRESHNESS_DECISION_OR_DENOMINATOR_MISMATCH")
	}
	wantSummary := model.Summary{
		CasesObserved: model.CaseTotal, CurrentEvidenceCases: model.CurrentEvidenceTotal,
		SyntheticCounterexamples: model.SyntheticCounterexampleTotal, AxisChangesObserved: model.AxisTotal,
		FixedAxisDenominator: model.AxisTotal, RawFreshCases: 7, RawStaleCases: 2, RawUnknownCases: 1,
		SemanticFreshCases: 8, SemanticStaleCases: 1, SemanticUnknownCases: 1,
		ClaimFreshCases: 2, ClaimStaleCases: 7, ClaimUnknownCases: 1,
		RawStaleByStage: map[string]int{model.StageMaterial: 2},
		StaleByStage: map[string]int{model.StageSubject: 1, model.StageMaterial: 1, model.StageRecipe: 1,
			model.StageEnvironment: 1, model.StageRunner: 1, model.StageVerifier: 2},
		UnknownByStage: map[string]int{model.StageSubject: 1}, FreshnessTransitions: model.TransitionTotal,
		ClaimLedgerEntries: model.TransitionTotal, ClaimDischarged: 2, ClaimOpenPreserved: 8, ClaimRefuted: 0,
		SourceReconstructedCases: 9, SourceUnavailableCases: 1,
		ForbiddenDependencyCount: 0,
		IndependenceContract:     model.FixedMetric{Numerator: model.IndependenceContractTotal, Denominator: model.IndependenceContractTotal},
		ReadOnlyBeforeCount:      0, ReadOnlyAfterCount: 0, ReadOnlyWriteSetStable: true,
	}
	if !reflect.DeepEqual(report.Summary, wantSummary) {
		return fmt.Errorf("EVIDENCE_FRESHNESS_SUMMARY_MISMATCH")
	}
	current, comment, semantic, unavailable := caseByID(report.Cases, "current-exact-head"), caseByID(report.Cases, "synthetic-comment-only"), caseByID(report.Cases, "synthetic-semantic-change"), caseByID(report.Cases, "synthetic-source-unavailable")
	if current == nil || current.ObservationClass != model.ObservationCurrent || !current.SourceAvailable || current.ObservedState != model.StateFresh || current.ObservedDecision != model.DecisionPass || current.RawFreshness != model.StateFresh || current.SemanticFreshness != model.StateFresh ||
		comment == nil || comment.ObservationClass != model.ObservationSynthetic || comment.RawFreshness != model.StateStale || comment.SemanticFreshness != model.StateFresh || comment.ObservedState != model.StateFresh || comment.ObservedDecision != model.DecisionPass ||
		semantic == nil || semantic.RawFreshness != model.StateStale || semantic.SemanticFreshness != model.StateStale || semantic.ObservedState != model.StateStale || semantic.ObservedDecision != model.DecisionFailClosed ||
		unavailable == nil || unavailable.SourceAvailable || unavailable.ObservedState != model.StateUnknown || unavailable.ObservedDecision != model.DecisionFailClosed || unavailable.ObservedResolution != model.ResolutionLower {
		return fmt.Errorf("EVIDENCE_FRESHNESS_INTERVENTION_MISMATCH")
	}
	previous := ""
	for index, item := range report.Cases {
		if item.Status != "OBSERVED" || item.Transition.From != "CLAIM_JUSTIFIED" || item.Transition.EvidenceDigest != report.ReceiptDigest || len(item.Checks) != model.CheckTotal {
			return fmt.Errorf("EVIDENCE_FRESHNESS_CASE_MISMATCH_%d", index)
		}
		entry := report.ClaimLedger[index]
		if entry.Schema != model.LedgerSchema || entry.Sequence != index+1 || entry.PriorState != model.ClaimOpen || entry.ClaimID != item.Transition.ClaimID ||
			entry.FreshnessObservation != item.ObservedState || entry.ReceiptDigest != report.ReceiptDigest || entry.PreviousDigest != previous || !model.VerifyLedgerEntryDigest(entry) || len(entry.Provenance) < 4 {
			return fmt.Errorf("EVIDENCE_FRESHNESS_LEDGER_MISMATCH_%d", index)
		}
		if item.ObservedState == model.StateFresh && entry.NextState != model.ClaimDischarged {
			return fmt.Errorf("EVIDENCE_FRESHNESS_DISCHARGE_MISMATCH_%d", index)
		}
		if (item.ObservedState == model.StateStale || item.ObservedState == model.StateUnknown) && entry.NextState != model.ClaimOpen {
			return fmt.Errorf("EVIDENCE_FRESHNESS_OPEN_PRESERVATION_MISMATCH_%d", index)
		}
		previous = entry.Digest
	}
	if report.ClaimLedgerDigest != model.DigestJSON(report.ClaimLedger) {
		return fmt.Errorf("EVIDENCE_FRESHNESS_LEDGER_DIGEST_MISMATCH")
	}
	for index, definition := range contract.Metrics {
		item := report.Indicators[index]
		if item.MetricID != definition.MetricID || item.Class != definition.Class || item.Producer != definition.Producer ||
			item.Consumer != definition.Consumer || item.ProofChoice != definition.ProofChoice || item.MetaOperation != definition.MetaOperation ||
			item.Numerator != definition.ExpectedNumerator || item.Denominator != definition.Denominator || item.ExpectedNumerator != definition.ExpectedNumerator ||
			!item.Satisfied || item.BasisPoints != 10000 {
			return fmt.Errorf("EVIDENCE_FRESHNESS_INDICATOR_MISMATCH_%d", index)
		}
	}
	if !reflect.DeepEqual(report.NotClaimed, contract.NotClaimed) || report.Digest != reportDigest(report) {
		return fmt.Errorf("EVIDENCE_FRESHNESS_DIGEST_OR_CLAIM_MISMATCH")
	}
	return nil
}

func validWriteSet(writeSet model.WriteSetObservation) bool {
	return model.ValidDigest(writeSet.BeforeDigest) && model.ValidDigest(writeSet.AfterDigest) &&
		writeSet.BeforeDigest == writeSet.AfterDigest && writeSet.BeforeCount == 0 && writeSet.AfterCount == 0
}

func caseByID(cases []model.CaseResult, id string) *model.CaseResult {
	for index := range cases {
		if cases[index].ID == id {
			return &cases[index]
		}
	}
	return nil
}
