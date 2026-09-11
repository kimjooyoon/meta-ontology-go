package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicresolutionrepair"
)

func verify(reportPath, humanPath string) error {
	if reportPath == "" || humanPath == "" {
		return errors.New("report and human-output are required")
	}
	value, err := decodeReport(reportPath)
	if err != nil {
		return err
	}
	checks := []func(report) error{
		validateReportIdentity,
		validateCounterexampleAndOverlay,
		validateCaseBindings,
		validateFallbackCase,
		validateRepairedCase,
		validateCaseCausalStates,
	}
	for _, check := range checks {
		if err := check(value); err != nil {
			return err
		}
	}
	return writeNew(humanPath, []byte(humanReport(value)), 0o444)
}

func decodeReport(reportPath string) (report, error) {
	data, err := readRegular(reportPath)
	if err != nil {
		return report{}, err
	}
	var value report
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return report{}, err
	}
	return value, nil
}

func validateReportIdentity(value report) error {
	if value.Schema != "gooo/public-semantic-resolution-repair-verification/v1" || value.Decision != publicresolutionrepair.DecisionClosed || value.CaseDenominator != publicresolutionrepair.CaseCount || value.ClosedCases != 2 || value.UnknownCases != 2 || value.RefutedCases != 2 || value.ResolutionLevelCount != publicresolutionrepair.ResolutionLevelCount || value.ProofModeObservationCount != publicresolutionrepair.ProofObservationCount || value.ProofFoundationCount != publicresolutionrepair.ProofFoundationCount || value.ProofCoherenceCount != publicresolutionrepair.ProofCoherenceCount || value.ProofRegressionCount != publicresolutionrepair.ProofRegressionCount || value.RepairProposalCount != publicresolutionrepair.RepairProposalCount || value.AuthorizationDecisionCount != publicresolutionrepair.AuthorizationDecisionCount || value.GraphEdgesBefore != publicresolutionrepair.GraphEdgeCountBefore || value.GraphEdgesAfter != publicresolutionrepair.GraphEdgeCountAfter || value.CanonicalGraphEdgeCount != publicresolutionrepair.CanonicalGraphEdgeCount || value.TestUnitsTotal != publicresolutionrepair.TestUnitCount || value.FallbackTestUnitsExecuted != publicresolutionrepair.FallbackExecuted || value.FallbackTestUnitsReused != publicresolutionrepair.FallbackReused || value.OverlayTestUnitsExecuted != publicresolutionrepair.OverlayExecuted || value.OverlayTestUnitsReused != publicresolutionrepair.OverlayReused || value.SelectivityTestUnitsExecuted != publicresolutionrepair.SelectivityExecuted || value.SelectivityTestUnitsReused != publicresolutionrepair.SelectivityReused || value.ContinuityEdgeCount != publicresolutionrepair.ContinuityEdgeCount || value.GeneratedArtifactCount != publicresolutionrepair.GeneratedArtifactCount || value.EvidenceArtifactCount != publicresolutionrepair.EvidenceArtifactCount || value.RuntimeComparable || value.RuntimeUnknown != "RUNTIME_MODES_NOT_EQUIVALENT" || value.RepositoryWrites != 0 || value.LocalTestExecutions != 0 || len(value.PublishedArtifacts) != publicresolutionrepair.EvidenceArtifactCount || !sameStrings(value.PublishedArtifacts, publicationNames) {
		return errors.New("semantic resolution repair report denominator or safety contract is invalid")
	}
	if err := value.Policy.Validate(); err != nil {
		return err
	}
	return nil
}

func validateCounterexampleAndOverlay(value report) error {
	if !value.OriginalCounterexample.Valid || value.OriginalCounterexample.CaseID != publicresolutionrepair.OriginalCounterexampleCaseID || value.OriginalCounterexample.Decision != publicresolutionrepair.DecisionRefuted || value.OriginalCounterexample.Reason != "FAIL_CLOSED_PARTIAL_REUSE_CONTRADICTION" || value.Proposal.From != value.OriginalCounterexample.ChangedComponent || value.Proposal.To != value.OriginalCounterexample.OmittedTarget || value.Proposal.ProofMode != publicresolutionrepair.ProofRegression || value.Proposal.Digest == "" {
		return errors.New("original v15 REFUTED counterexample or deterministic proposal was not preserved")
	}
	if err := publicresolutionrepair.ValidateAuthorization(value.Authorization, value.Proposal); err != nil {
		return err
	}
	if err := publicresolutionrepair.ValidateOverlay(value.Overlay, value.Policy, value.Proposal, value.Authorization); err != nil {
		return err
	}
	return nil
}

func validateCaseBindings(value report) error {
	if len(value.Cases) != len(value.Policy.Cases) {
		return errors.New("semantic repair case table is incomplete")
	}
	for index, item := range value.Cases {
		policyCase := value.Policy.Cases[index]
		if item.CaseID != policyCase.ID || item.ExpectedDecision != policyCase.Decision || item.ProofMode != policyCase.ProofMode || item.ResolutionBefore != policyCase.ResolutionFrom || item.ResolutionAfter != policyCase.ResolutionTo {
			return fmt.Errorf("semantic repair case %d is not bound to canonical .gooo policy", index+1)
		}
		if err := publicresolutionrepair.CompareCase(item, item.ExpectedDecision); err != nil {
			return err
		}
	}
	return nil
}

func validateFallbackCase(value report) error {
	fallback := value.Cases[0]
	if fallback.Decision != publicresolutionrepair.DecisionClosed || fallback.ResolutionBefore != publicresolutionrepair.ResolutionSelective || fallback.ResolutionAfter != publicresolutionrepair.ResolutionFallback || fallback.ProofMode != publicresolutionrepair.ProofRegression || !fallback.OriginalCounterexamplePreserved || fallback.Fallback.TestUnitsExecuted != 2 || fallback.Fallback.TestUnitsReused != 0 || fallback.Fallback.TestExecutions != 2 || fallback.ImpactedClosureBefore != 1 || fallback.ImpactedClosureAfter != 1 || fallback.FalseNegativeImpactedTestsBefore != 1 || fallback.FalseNegativeImpactedTestsAfter != 1 {
		return errors.New("full-project fallback does not conservatively preserve the original hidden-dependency counterexample")
	}
	return nil
}

func validateRepairedCase(value report) error {
	repaired := value.Cases[1]
	if repaired.Decision != publicresolutionrepair.DecisionClosed || repaired.ResolutionBefore != publicresolutionrepair.ResolutionFallback || repaired.ResolutionAfter != publicresolutionrepair.ResolutionOverlay || repaired.ProofMode != publicresolutionrepair.ProofRegression || repaired.GraphEdgesBefore != 1 || repaired.GraphEdgesAfter != 2 || repaired.ProposedEdges != 1 || repaired.AuthorizedEdges != 1 || repaired.ImpactedClosureBefore != 1 || repaired.ImpactedClosureAfter != 2 || repaired.FalseNegativeImpactedTestsBefore != 1 || repaired.FalseNegativeImpactedTestsAfter != 0 || repaired.OverlayReplay.TestUnitsExecuted != 2 || repaired.OverlayReplay.TestUnitsReused != 0 || repaired.UnchangedPartitionSelectivity.TestUnitsExecuted != 1 || repaired.UnchangedPartitionSelectivity.TestUnitsReused != 1 || !repaired.UnchangedSelectivityProven || !repaired.SafetyImprovement {
		return errors.New("authorized overlay replay does not prove complete closure and unaffected selectivity")
	}
	if !repaired.Comparisons.GeneratedBytesEqual || !repaired.Comparisons.GeneratedSemanticEqual || !repaired.Comparisons.TestContractEqual || !repaired.Comparisons.FullTestOutcomeEqual || !repaired.Comparisons.OverlayBindingEqual || !value.GeneratedBytesEqual || !value.SemanticEqual || !value.TestContractEqual || !value.FullTestOutcomeEqual || !value.OverlayBindingEqual || !value.SafetyImprovement {
		return errors.New("authorized repair does not prove equivalent full outcomes and safety improvement")
	}
	return nil
}

func validateCaseCausalStates(value report) error {
	for _, item := range value.Cases {
		if item.Decision == publicresolutionrepair.DecisionUnknown && (item.Unknown == nil || item.Unknown.Stage == "" || item.Unknown.Step == "" || item.Unknown.Reason == "" || item.Unknown.UnknownClass == "" || item.Unknown.NextOperation == "" || len(item.Unknown.BlockedBy) == 0) {
			return errors.New("UNKNOWN repair case is missing one or more causal fields")
		}
		if item.Decision == publicresolutionrepair.DecisionRefuted && item.Unknown != nil {
			return errors.New("REFUTED repair case did not dominate UNKNOWN")
		}
	}
	return nil
}

func publish(root string, value report) error {
	paths := map[string]string{
		"canonical-source.gooo": filepathJoin(root, "canonical-source.gooo"), "canonical-test.go": filepathJoin(root, "canonical-test.go"), "upstream-orchestration-report.json": filepathJoin(root, "upstream-orchestration-report.json"),
		"original-hidden-dependency.json": filepathJoin(root, "original-hidden-dependency.json"), "original-partial-reuse-report.json": filepathJoin(root, "original-partial-reuse-report.json"), "v15-counterexample-provenance.json": filepathJoin(root, "v15-counterexample-provenance.json"),
		"resolution-repair-policy.json": filepathJoin(root, "resolution-repair-policy.json"), "fallback-generated.go": filepathJoin(root, "fallback-generated.go"), "fallback-generated.manifest.jsonl": filepathJoin(root, "fallback-generated.manifest.jsonl"),
		"fallback-baseline.json": filepathJoin(root, "fallback-baseline.json"), "fallback-result.json": filepathJoin(root, "fallback-result.json"), "fallback-counterexample-preserved.json": filepathJoin(root, "fallback-counterexample-preserved.json"),
		"repair-proposal.json": filepathJoin(root, "repair-proposal.json"), "repair-authorization.json": filepathJoin(root, "repair-authorization.json"), "authorized-graph-overlay.json": filepathJoin(root, "authorized-graph-overlay.json"),
		"overlay-generated.go": filepathJoin(root, "overlay-generated.go"), "overlay-generated.manifest.jsonl": filepathJoin(root, "overlay-generated.manifest.jsonl"), "overlay-replay.json": filepathJoin(root, "overlay-replay.json"),
		"overlay-selectivity.json": filepathJoin(root, "overlay-selectivity.json"), "overlay-outcome-comparison.json": filepathJoin(root, "overlay-outcome-comparison.json"), "ambiguous-repair-evidence.json": filepathJoin(root, "ambiguous-repair-evidence.json"),
		"unsupported-repair-proof-mode.json": filepathJoin(root, "unsupported-repair-proof-mode.json"), "tampered-counterexample.json": filepathJoin(root, "tampered-counterexample.json"), "unauthorized-repair.json": filepathJoin(root, "unauthorized-repair.json"),
		"resolution-repair-report.json": filepathJoin(root, "resolution-repair-report.json"), "resolution-repair-human.txt": filepathJoin(root, "resolution-repair-human.txt"), "runtime-measurements.json": filepathJoin(root, "runtime-measurements.json"),
		"repository-status.json": filepathJoin(root, "repository-status.json"), "resolution-repair-verification-input.json": filepathJoin(root, "resolution-repair-verification-input.json"), "resolution-repair-case-table.json": filepathJoin(root, "resolution-repair-case-table.json"),
	}
	if len(value.PublishedArtifacts) != len(publicationNames) {
		return errors.New("semantic repair publication denominator is invalid")
	}
	for _, name := range publicationNames {
		filename, ok := paths[name]
		if !ok {
			return fmt.Errorf("semantic repair publication mapping is missing %s", name)
		}
		data, err := readRegular(filename)
		if err != nil {
			return err
		}
		if err := writeNew(filepathJoin(filepathJoin(root, "publish"), name), data, 0o444); err != nil {
			return err
		}
	}
	return nil
}

func humanReport(value report) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Semantic resolution repair\n\nDecision: `%s`\nReason: `%s`\n\n", value.Decision, value.Reason)
	fmt.Fprintf(&builder, "Cases: `%d CLOSED / %d UNKNOWN / %d REFUTED`\n", value.ClosedCases, value.UnknownCases, value.RefutedCases)
	fmt.Fprintf(&builder, "Resolution levels / proof observations / proposals / authorization decisions: `%d/%d/%d/%d`\n", value.ResolutionLevelCount, value.ProofModeObservationCount, value.RepairProposalCount, value.AuthorizationDecisionCount)
	fmt.Fprintf(&builder, "Graph edges before/after: `%d/%d`; impacted closure before/after: `%d/%d`; false-negative tests before/after: `%d/%d`\n", value.GraphEdgesBefore, value.GraphEdgesAfter, value.Cases[1].ImpactedClosureBefore, value.Cases[1].ImpactedClosureAfter, value.Cases[1].FalseNegativeImpactedTestsBefore, value.Cases[1].FalseNegativeImpactedTestsAfter)
	fmt.Fprintf(&builder, "Fallback tests executed/reused: `%d/%d`; overlay replay: `%d/%d`; unchanged selectivity: `%d/%d`\n", value.Fallback.TestUnitsExecuted, value.Fallback.TestUnitsReused, value.OverlayReplay.TestUnitsExecuted, value.OverlayReplay.TestUnitsReused, value.UnchangedSelectivity.TestUnitsExecuted, value.UnchangedSelectivity.TestUnitsReused)
	fmt.Fprintf(&builder, "Safety improvement: `%t`; generated/semantic/contract/outcome/overlay equal: `%t/%t/%t/%t/%t`\n", value.SafetyImprovement, value.GeneratedBytesEqual, value.SemanticEqual, value.TestContractEqual, value.FullTestOutcomeEqual, value.OverlayBindingEqual)
	fmt.Fprintf(&builder, "Repository writes / local test executions: `%d/%d`; runtime comparison: `UNKNOWN` (%s)\n", value.RepositoryWrites, value.LocalTestExecutions, value.RuntimeUnknown)
	return builder.String()
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func filepathJoin(left, right string) string {
	if strings.HasSuffix(left, "/") {
		return left + right
	}
	return left + "/" + right
}
