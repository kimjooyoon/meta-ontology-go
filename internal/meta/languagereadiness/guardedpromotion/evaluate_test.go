package guardedpromotion

import "testing"

func validSource() Source {
	predecessor := "5201f9ebc4bf67683ecab9fccf638560af0e2776"
	current := "75b892f4bc661b4a58a5517532c44662dae6eedf"
	return Source{
		RequestedRepository: "kimjooyoon/meta-ontology-go",
		ObservedRepository: "kimjooyoon/meta-ontology-go", DefaultBranch: "dev",
		CurrentHeadSHA: current, PredecessorSHA: predecessor,
		Workflow: WorkflowEvidence{
			RunID: 1, Name: "CI [push full]", Path: CIPath, Event: "push",
			Status: "completed", Conclusion: "success", HeadSHA: current, HeadBranch: "dev",
		},
		Artifact: ArtifactEvidence{
			RunID: 2, RunAttempt: 1, RunEvent: "workflow_run", ArtifactID: 3,
			ArtifactName: PromotionArtifactBase + predecessor,
			ArtifactDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			FileSHA256: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			ReportSchema: PromotionSchema,
			ReportDigest: "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			ReportCurrentHeadSHA: predecessor, ReportDecision: "PASS",
			ReportSatisfied: 8, ReportTotal: 8,
		},
		ObservedRuns: 1, ObservedArtifacts: 8, ValidCandidates: 1,
	}
}

func TestBuildAuthorizesMergedPush(t *testing.T) {
	report := Build(validSource())
	if report.Decision != DecisionAuthorized || report.Summary.ReadinessBPS != 10000 {
		t.Fatalf("decision=%s bps=%d", report.Decision, report.Summary.ReadinessBPS)
	}
	if report.Summary.Satisfied != 12 || !report.Summary.ReadinessPromotionAuthorized {
		t.Fatalf("summary=%+v", report.Summary)
	}
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
}

func TestBuildDeniesPullRequestWithoutHidingIt(t *testing.T) {
	source := validSource()
	source.Workflow.Name = "CI [PR authoritative]"
	source.Workflow.Event = "pull_request"
	source.Workflow.HeadBranch = "agent/example"
	report := Build(source)
	if report.Decision != DecisionDenied || report.Reason != ReasonMergedPushRequired {
		t.Fatalf("decision=%s reason=%s", report.Decision, report.Reason)
	}
	if report.Summary.Satisfied != 10 || report.Summary.ReadinessBPS != 8333 {
		t.Fatalf("summary=%+v", report.Summary)
	}
}
