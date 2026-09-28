package main

import "testing"

func TestBuildReportKeepsUnresolvedBoundaryVisible(t *testing.T) {
	input := inputDocument{
		Version:              "domain.capability.input.v1",
		Domain:               "support_triage",
		ExpectedCapabilities: []string{"declaration_analysis", "provenance"},
		Observations: []observation{
			{
				Domain:         "support_triage",
				CapabilityID:   "declaration_analysis",
				Outcome:        "useful",
				EvidenceDigest: "evidence-declaration-analysis",
				NonExecuting:   true,
				NonAuthorizing: true,
			},
			{
				Domain:         "support_triage",
				CapabilityID:   "provenance",
				Outcome:        "unresolved",
				EvidenceDigest: "evidence-provenance",
				NonExecuting:   true,
				NonAuthorizing: true,
			},
		},
	}

	report, err := buildReport(input)
	if err != nil {
		t.Fatalf("buildReport returned error: %v", err)
	}
	if report.CoverageRatio != 1 {
		t.Fatalf("coverage ratio = %v, want 1", report.CoverageRatio)
	}
	if report.UnresolvedRatio != 0.5 {
		t.Fatalf("unresolved ratio = %v, want 0.5", report.UnresolvedRatio)
	}
	if report.DecisionSignal != "COLLECT_MORE_OBSERVATIONS" {
		t.Fatalf("decision signal = %q, want COLLECT_MORE_OBSERVATIONS", report.DecisionSignal)
	}
	if report.MeasurementSemantics == "" {
		t.Fatal("measurement semantics must be explicit")
	}
}

func TestBuildReportRejectsAuthorizationCrossing(t *testing.T) {
	input := inputDocument{
		Version:              "domain.capability.input.v1",
		Domain:               "support_triage",
		ExpectedCapabilities: []string{"support_triage"},
		Observations: []observation{
			{
				Domain:         "support_triage",
				CapabilityID:   "support_triage",
				Outcome:        "useful",
				EvidenceDigest: "evidence-support-triage",
				NonExecuting:   false,
				NonAuthorizing: true,
			},
		},
	}

	if _, err := buildReport(input); err == nil {
		t.Fatal("buildReport accepted an executing observation")
	}
}
