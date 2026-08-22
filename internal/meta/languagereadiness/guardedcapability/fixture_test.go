package guardedcapability

import "testing"

func exactSource(t *testing.T) Source {
	t.Helper()
	report, err := foundationReport()
	if err != nil {
		t.Fatal(err)
	}
	return Source{
		CurrentHeadSHA: "6303fe037c0cdff15bd854953785fc968ca2c743",
		WorkflowRunID: FoundationWorkflowRunID, ArtifactID: FoundationArtifactID,
		ArtifactDigest: FoundationArtifactDigest, ReportFileSHA: FoundationReportFileSHA,
		FoundationReport: report, AncestryObserved: true, FoundationAncestor: true,
		GuardTreesObserved: true, FoundationGuardTree: "tree-a", CurrentGuardTree: "tree-a",
		WitnessTreesObserved: true, FoundationWitnessTree: "tree-b", CurrentWitnessTree: "tree-b",
	}
}
