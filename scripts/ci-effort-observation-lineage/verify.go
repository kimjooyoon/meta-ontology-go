package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicworkflowlineage"
)

func verifyReport(reportPath, humanPath string) error {
	if reportPath == "" {
		return errors.New("report is required")
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return err
	}
	var value report
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if err := validateConformance(value); err != nil {
		return err
	}
	if humanPath != "" {
		if err := os.WriteFile(humanPath, []byte(humanReport(value)), 0o444); err != nil {
			return err
		}
	}
	return nil
}

func validateConformance(value report) error {
	if value.Schema != publicworkflowlineage.ReportSchema || value.Decision != publicworkflowlineage.DecisionClosed || value.ProvenanceState != publicworkflowlineage.ProvenanceExact || value.CaseDenominator != publicworkflowlineage.CaseCount || value.ClosedCases != publicworkflowlineage.ClosedCaseCount || value.UnknownCases != publicworkflowlineage.UnknownCaseCount || value.RefutedCases != publicworkflowlineage.RefutedCaseCount || value.CasesSatisfied != publicworkflowlineage.CaseCount || value.CasesTotal != publicworkflowlineage.CaseCount || value.ActiveLineageRoots != 1 || value.LineageEdgeCount != publicworkflowlineage.LineageEdgeCount || value.SourceReceiptCount != publicworkflowlineage.SourceReceiptCount || value.ConsumerReceiptCount != publicworkflowlineage.ConsumerReceiptCount || value.EvidenceArtifactCount != publicworkflowlineage.EvidenceArtifactCount || value.StaleMisattributedBefore != 2 || value.StaleMisattributedAfter != 0 || value.ExactSubjectBindings != 3 || value.UnknownClassifications != 3 || value.StaleSourceStatesUnknown != 3 || value.UnknownSixFieldPreservations != 3 || value.ContradictionsRefuted != 3 || value.MismatchDetections != 3 || value.ExactReplayComparisons != 1 || value.FallbackAttempts != 1 || value.FallbackAccepted != 0 || value.FallbackRejected != 1 || value.SourceArtifactResolutions != 3 || !value.TrueProductFailuresNotRelabeled || value.WallMS <= 0 || value.PeakRSSKib <= 0 || value.RuntimeComparable || value.RuntimeUnknown != "RUNTIME_MODES_NOT_EQUIVALENT" || value.RepositoryWrites != 0 || value.LocalTestExecutions != 0 || len(value.PublishedArtifacts) != publicworkflowlineage.EvidenceArtifactCount {
		return fmt.Errorf("workflow lineage report denominator or safety contract is invalid: decision=%q provenance=%q cases=%d/%d/%d/%d satisfied=%d/%d roots=%d artifacts=%d exact=%d unknown=%d stale_unknown=%d unknown_fields=%d contradictions=%d mismatch=%d replay=%d fallback=%d/%d/%d resolved=%d published=%d", value.Decision, value.ProvenanceState, value.CaseDenominator, value.ClosedCases, value.UnknownCases, value.RefutedCases, value.CasesSatisfied, value.CasesTotal, value.ActiveLineageRoots, value.EvidenceArtifactCount, value.ExactSubjectBindings, value.UnknownClassifications, value.StaleSourceStatesUnknown, value.UnknownSixFieldPreservations, value.ContradictionsRefuted, value.MismatchDetections, value.ExactReplayComparisons, value.FallbackAttempts, value.FallbackAccepted, value.FallbackRejected, value.SourceArtifactResolutions, len(value.PublishedArtifacts))
	}
	if err := value.Policy.Validate(); err != nil {
		return err
	}
	if !sameStrings(value.PublishedArtifacts, publicationNames) {
		return errors.New("workflow lineage publication list is not fixed")
	}
	if len(value.Cases) != len(value.Policy.Cases) {
		return errors.New("workflow lineage case table is incomplete")
	}
	for index, item := range value.Cases {
		if item.CaseID != value.Policy.Cases[index].ID || item.ExpectedDecision != value.Policy.Cases[index].Decision || item.Decision != item.ExpectedDecision || item.LineageState != value.Policy.Cases[index].LineageState {
			return fmt.Errorf("workflow lineage case %d is not bound to canonical policy", index+1)
		}
		if item.Decision == publicworkflowlineage.DecisionUnknown && !causalComplete(item.Unknown) {
			return fmt.Errorf("UNKNOWN case %s is missing causal fields", item.CaseID)
		}
		if item.Decision == publicworkflowlineage.DecisionRefuted && item.Unknown != nil {
			return fmt.Errorf("REFUTED case %s carries UNKNOWN evidence", item.CaseID)
		}
	}
	return nil
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
