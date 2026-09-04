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
	if value.Schema != publicworkflowlineage.ReportSchema || value.Decision != publicworkflowlineage.DecisionClosed || value.CaseDenominator != publicworkflowlineage.CaseCount || value.ClosedCases != publicworkflowlineage.ClosedCaseCount || value.UnknownCases != publicworkflowlineage.UnknownCaseCount || value.RefutedCases != publicworkflowlineage.RefutedCaseCount || value.LineageEdgeCount != publicworkflowlineage.LineageEdgeCount || value.SourceReceiptCount != publicworkflowlineage.SourceReceiptCount || value.ConsumerReceiptCount != publicworkflowlineage.ConsumerReceiptCount || value.EvidenceArtifactCount != publicworkflowlineage.EvidenceArtifactCount || value.StaleMisattributedBefore != 2 || value.StaleMisattributedAfter != 0 || value.ExactSubjectBindings != 2 || value.UnknownClassifications != 2 || value.StaleSourceStatesUnknown != 2 || value.MismatchDetections != 2 || value.FallbackAttempts != 1 || value.FallbackAccepted != 0 || value.FallbackRejected != 1 || value.SourceArtifactResolutions != 2 || !value.TrueProductFailuresNotRelabeled || value.WallMS <= 0 || value.PeakRSSKib <= 0 || value.RuntimeComparable || value.RuntimeUnknown != "RUNTIME_MODES_NOT_EQUIVALENT" || value.RepositoryWrites != 0 || value.LocalTestExecutions != 0 || len(value.PublishedArtifacts) != publicworkflowlineage.EvidenceArtifactCount {
		return errors.New("workflow lineage report denominator or safety contract is invalid")
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
		if item.CaseID != value.Policy.Cases[index].ID || item.ExpectedDecision != value.Policy.Cases[index].Decision || item.Decision != item.ExpectedDecision {
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
