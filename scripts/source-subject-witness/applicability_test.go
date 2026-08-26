package main

import "testing"

func TestWorkflowDiscoveryApplicabilityIsBound(t *testing.T) {
	directory, rows := workflowDiscoveryFixture()
	binding, err := storageDirectoryBinding(directory, indexIndicators(rows), 0)
	if err != nil {
		t.Fatal(err)
	}
	if binding.IndicatorCount != 6 || binding.ApplicableIndicators != 5 || binding.NotApplicableIndicators != 1 {
		t.Fatalf("binding applicability counts = %d/%d/%d", binding.IndicatorCount, binding.ApplicableIndicators, binding.NotApplicableIndicators)
	}
	if binding.Operation != "observe+preserve-workflow-discovery+separate-directory-kinds" {
		t.Fatalf("binding operation = %q", binding.Operation)
	}
}

func TestWorkflowDiscoveryApplicabilityRejectsUnknownRule(t *testing.T) {
	directory, rows := workflowDiscoveryFixture()
	rows[0].ApplicabilityRuleID = "unknown"
	if _, err := storageDirectoryBinding(directory, indexIndicators(rows), 0); err == nil {
		t.Fatal("unknown workflow discovery rule was accepted")
	}
}

func TestWorkflowDiscoveryApplicabilityRejectsOtherDirectory(t *testing.T) {
	directory, rows := workflowDiscoveryFixture()
	directory.Path = "other"
	for index := range rows {
		rows[index].Subject = directory.Path
	}
	if _, err := storageDirectoryBinding(directory, indexIndicators(rows), 0); err == nil {
		t.Fatal("non-workflow topology exemption was accepted")
	}
}
