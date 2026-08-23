package guardedpromotion

import "testing"

func TestUnknownEventLowersResolution(t *testing.T) {
	source := validSource()
	source.Workflow.Name = ""
	source.Workflow.Event = "mystery"
	report := Build(source)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionLower {
		t.Fatalf("decision=%s resolution=%s", report.Decision, report.Resolution)
	}
	if report.Summary.Unresolved == 0 {
		t.Fatal("unknown event did not create unresolved evidence")
	}
}

func TestRepositoryMismatchLowersResolution(t *testing.T) {
	source := validSource()
	source.ObservedRepository = "attacker/example"
	report := Build(source)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionLower {
		t.Fatalf("decision=%s resolution=%s", report.Decision, report.Resolution)
	}
	if report.Reason != ReasonRepositoryMismatch {
		t.Fatalf("reason=%s", report.Reason)
	}
}

func TestAmbiguousPredecessorFailsClosed(t *testing.T) {
	source := validSource()
	source.ValidCandidates = 2
	source.AmbiguousCandidates = 2
	report := Build(source)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionLower {
		t.Fatalf("decision=%s resolution=%s", report.Decision, report.Resolution)
	}
}

func TestMutationAuthorityIsDenied(t *testing.T) {
	source := validSource()
	source.RepositoryMutationAuthorized = true
	report := Build(source)
	if report.Decision != DecisionDenied || report.Summary.ReadinessPromotionAuthorized {
		t.Fatalf("decision=%s summary=%+v", report.Decision, report.Summary)
	}
}

func TestValidateRejectsTampering(t *testing.T) {
	report := Build(validSource())
	report.Summary.Satisfied--
	if err := Validate(report); err == nil {
		t.Fatal("tampered report was accepted")
	}
}
