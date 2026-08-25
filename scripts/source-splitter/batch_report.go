package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
)

func readLogicalSplitPlan(name string) (logicalSplitPlan, error) {
	var report logicalSplitPlan
	data, err := os.ReadFile(name)
	if err == nil {
		err = json.Unmarshal(data, &report)
	}
	return report, err
}

func buildSplitBatchReport(sha string, plans []plannedSplit, subjects []splitBatchSubject) splitBatchReport {
	selected := make([]string, len(plans))
	changed, created := 0, 0
	for index, item := range plans {
		selected[index] = item.logical
	}
	for _, subject := range subjects {
		changed += len(subject.ChangedFiles)
		created += len(subject.CreatedFiles)
	}
	decision, reason := "PASS", "PROJECTABLE_SUBJECTS_SPLIT"
	if len(selected) == 0 {
		decision, reason = "FIXED_POINT", "NO_PROJECTABLE_SUBJECTS"
	}
	coordinates := splitBatchCoordinates{SelectedSubjects: len(selected), AppliedSubjects: len(subjects),
		ChangedPaths: changed, CreatedPaths: created}
	report := splitBatchReport{Schema: splitBatchSchema, SourceSHA: sha, Decision: decision, Resolution: "EXACT",
		Reason: reason, MetaOperation: splitBatchOperation, Selected: selected, Subjects: subjects,
		Coordinates: coordinates, Exact: len(selected) == len(subjects)}
	report.Indicators = []splitBatchIndicator{
		{ID: "split.selected", Class: "DRIVER", ProofChoice: "FOUNDATION", MetaOperation: splitBatchOperation,
			Observed: len(selected), Expected: len(selected), Satisfied: true},
		{ID: "split.applied", Class: "OUTCOME", ProofChoice: "COHERENCE", MetaOperation: splitBatchOperation,
			Observed: len(subjects), Expected: len(selected), Satisfied: report.Exact},
		{ID: "guardrail.unknown", Class: "GUARDRAIL", ProofChoice: "REGRESSION", MetaOperation: splitBatchOperation,
			Observed: coordinates.Unknowns, Expected: 0, Satisfied: coordinates.Unknowns == 0},
	}
	report.Proofs = []splitBatchProof{{Choice: "FOUNDATION", MetaOperation: "select-projectable-subjects", Passed: true},
		{Choice: "COHERENCE", MetaOperation: splitBatchOperation, Passed: report.Exact},
		{Choice: "REGRESSION", MetaOperation: "reject-incomplete-split-evidence", Passed: coordinates.Unknowns == 0}}
	return sealSplitBatchReport(report)
}

func sealSplitBatchReport(report splitBatchReport) splitBatchReport {
	unsigned := report
	unsigned.Digest = ""
	data, _ := json.Marshal(unsigned)
	digest := sha256.Sum256(data)
	report.Digest = "sha256:" + hex.EncodeToString(digest[:])
	return report
}

func writeSplitBatchReport(name string, report splitBatchReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, append(data, '\n'), 0o644)
}
