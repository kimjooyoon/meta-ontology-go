package judge

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Indicators(summary model.Summary) []model.Indicator {
	return []model.Indicator{
		indicator("gooo.metric.bounded-transformation.source-derived-cases.v2", model.ProofFoundation, "derive-source-case-inventory", summary.SourceDerivedCases, summary.BoundedInputDomainDenominator, "="),
		indicator("gooo.metric.bounded-transformation.unique-claim-instances.v2", model.ProofCoherence, "case-qualified-claim-ledger", summary.UniqueClaimInstances, summary.CasesTotal*len(model.CanonicalValueSpecs()), "="),
		indicator("gooo.metric.bounded-transformation.accepted-transitions.v2", model.ProofCoherence, "verify-transition-digests", summary.AcceptedTransitions, summary.UniqueClaimInstances, "="),
		indicator("gooo.metric.bounded-transformation.domain-observations.v2", model.ProofFoundation, "observe-bounded-input-domain", summary.BoundedInputDomainObservations, summary.BoundedInputDomainDenominator, "="),
		indicator("gooo.metric.bounded-transformation.provisional-receipts.v2", model.ProofFoundation, "emit-provisional-receipts", summary.ProvisionalReceipts, summary.SourceDerivedCases, "="),
		indicator("gooo.metric.bounded-transformation.authorization-receipts.v2", model.ProofCoherence, "independent-authorization", summary.AuthorizationReceipts, 2, "="),
		indicator("gooo.metric.bounded-transformation.executed-effects.v2", model.ProofCoherence, "post-judgment-effect-gate", summary.ExecutedEffects, 1, "="),
		indicator("gooo.metric.bounded-transformation.independently-observed-effects.v2", model.ProofCoherence, "read-only-effect-observation", summary.IndependentlyObservedEffects, 1, "="),
		indicator("gooo.metric.bounded-transformation.unknown-effect-scopes.v2", model.ProofFoundation, "separate-transient-write-scope", summary.UnknownEffectScopes, 1, "="),
		indicator("gooo.metric.bounded-transformation.corrections.v2", model.ProofFoundation, "fixed-correction-denominator", summary.CorrectionCount, summary.CorrectionDenominator, "="),
	}
}

func indicator(id, proof, operation string, value, target int, relation string) model.Indicator {
	return model.Indicator{MetricID: id, Producer: model.ProducerID, Consumer: model.ConsumerID, MetaOperation: operation, ProofChoice: proof, Value: value, Target: target, Relation: relation, Satisfied: value == target}
}

// ValidateReport is a report consumer check. It independently judges every
// receipt and compares only against labeled validator expectations; it does
// not call producer.Build or treat deterministic replay as evidence.
func ValidateReport(report model.Report, source []byte) error {
	if report.Schema != model.ReportSchema || !model.ValidHead(report.HeadSHA) || report.SourcePath != model.SourcePath ||
		report.SourceDigest != model.DigestBytes(source) || report.ContractDigest != model.ValueContractDigest() ||
		report.ValidatorContractDigest != model.ValidatorContractDigest() || report.DenominatorID != model.DenominatorID || report.DenominatorTotal != len(report.Cases) ||
		report.Digest == "" || report.Digest != model.SealReport(report).Digest {
		return fmt.Errorf("report identity or digest is invalid")
	}
	ids, err := sourceCaseIDs(source)
	if err != nil || len(ids) != len(report.Cases) {
		return fmt.Errorf("report case inventory is not source-derived")
	}
	seen := map[string]bool{}
	for _, result := range report.Cases {
		if seen[result.Receipt.CaseID] {
			return fmt.Errorf("duplicate report case %q", result.Receipt.CaseID)
		}
		seen[result.Receipt.CaseID] = true
		if !contains(ids, result.Receipt.CaseID) {
			return fmt.Errorf("report case %q is not source-derived", result.Receipt.CaseID)
		}
		judgment := Judge(result.Receipt, source)
		if !judgment.Independent || !reflect.DeepEqual(judgment, result.Judgment) {
			return fmt.Errorf("case %q independent judgment mismatch", result.Receipt.CaseID)
		}
		if !model.ValidDigest(result.ProvisionalReceiptDigest) || (result.Receipt.Phase == model.ReceiptExecuted && result.ProvisionalReceiptDigest == result.Receipt.Digest) ||
			(result.Judgment.Decision == model.DecisionAllowed && !model.ValidDigest(result.AuthorizationReceiptDigest)) ||
			(result.Judgment.Decision != model.DecisionAllowed && result.AuthorizationReceiptDigest != "") || result.ExecutedEffects != len(result.Receipt.Effects) {
			return fmt.Errorf("case %q receipt phase metrics are not bound", result.Receipt.CaseID)
		}
		expectation, ok := model.ValidatorExpectationFor(model.CanonicalContract(), result.Receipt.CaseID)
		if !ok || result.Expectation != expectation || result.Judgment.Decision != expectation.ExpectedDecision || result.Judgment.Resolution != expectation.ExpectedResolution ||
			result.Judgment.Reason != expectation.ExpectedReason || result.Judgment.Status != expectation.ExpectedStatus || len(result.Receipt.Effects) != expectation.ExpectedEffectCount || !result.Satisfied {
			return fmt.Errorf("case %q does not satisfy labeled validator expectation", result.Receipt.CaseID)
		}
	}
	if report.Decision != model.DecisionPass || report.Resolution != model.ResolutionExact || report.Reason != "ALL_BOUNDED_CASES_SATISFIED" {
		return fmt.Errorf("report top decision is not the derived bounded witness result")
	}
	if report.Summary.CorrectionCount != 12 || report.Summary.CorrectionDenominator != 12 || len(report.Indicators) != len(Indicators(report.Summary)) ||
		!reflect.DeepEqual(report.Indicators, Indicators(report.Summary)) {
		return fmt.Errorf("report metrics or correction denominator are invalid")
	}
	return nil
}

func sourceCaseIDs(source []byte) ([]string, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("source syntax: %s", diagnostics.Error())
	}
	ids := []string{}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || len(activity.Parameters) != 0 || activity.Result.Name != "Transformation" || !activity.ValueProgramPresent {
			continue
		}
		fields := stringsSplitFields(activity.ValueProgram)
		if fields["case"] != "" {
			ids = append(ids, fields["case"])
		}
	}
	return ids, nil
}

func stringsSplitFields(program string) map[string]string {
	fields := map[string]string{}
	for _, part := range splitSemicolon(program) {
		key, value, ok := cutEqual(part)
		if ok {
			fields[key] = value
		}
	}
	return fields
}

// Small source-only helpers keep report validation independent of producer.
func splitSemicolon(value string) []string         { return strings.Split(value, ";") }
func cutEqual(value string) (string, string, bool) { return strings.Cut(value, "=") }
func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
