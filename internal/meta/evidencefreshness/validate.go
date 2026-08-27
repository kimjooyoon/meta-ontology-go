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
		report.ReceiptDigest != report.Receipt.Digest || report.Independence.Schema != model.IndependenceSchema ||
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
		PreservationTransitions: model.TransitionTotal, TemporalBoundaryCases: 1, ReadOnlyCases: 1}
	if !reflect.DeepEqual(report.Summary, wantSummary) {
		return fmt.Errorf("EVIDENCE_FRESHNESS_SUMMARY_MISMATCH")
	}
	for index, item := range report.Cases {
		if item.Status != "SATISFIED" || len(item.Checks) != model.CheckTotal || item.Transition.From != "CLAIM_JUSTIFIED" || item.Transition.Coordinate.Stage == "" {
			return fmt.Errorf("EVIDENCE_FRESHNESS_CASE_MISMATCH_%d", index)
		}
	}
	for index, definition := range contract.Metrics {
		item := report.Indicators[index]
		if item.MetricID != definition.MetricID || item.Class != definition.Class || item.Producer != definition.Producer ||
			item.Consumer != definition.Consumer || item.ProofChoice != definition.ProofChoice || item.MetaOperation != definition.MetaOperation ||
			item.Denominator != definition.Denominator || item.ExpectedNumerator != definition.ExpectedNumerator || !item.Satisfied ||
			item.BasisPoints != 10000 {
			return fmt.Errorf("EVIDENCE_FRESHNESS_INDICATOR_MISMATCH_%d", index)
		}
	}
	if !reflect.DeepEqual(report.NotClaimed, contract.NotClaimed) || report.Digest != reportDigest(report) {
		return fmt.Errorf("EVIDENCE_FRESHNESS_DIGEST_OR_CLAIM_MISMATCH")
	}
	return nil
}
