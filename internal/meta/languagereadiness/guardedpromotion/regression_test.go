package guardedpromotion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestPromotionEvidenceWorkflowSeparatesProducerFromJudge(t *testing.T) {
	workflowPath := filepath.Join("..", "..", "..", "..", TransformationPath)
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(raw)
	for _, fragment := range []string{
		"name: Metric counterfactual conformance",
		"actions: read",
	} {
		if !strings.Contains(workflow, fragment) {
			t.Fatalf("promotion evidence authority is missing %q", fragment)
		}
	}
	start := strings.Index(workflow, "\n  proposal-promotion-evidence:\n")
	if start < 0 {
		t.Fatal("promotion evidence producer job is missing")
	}
	producerTail := workflow[start+1:]
	producer, _, ok := strings.Cut(producerTail, "\n  program:\n")
	if !ok {
		t.Fatal("promotion evidence producer boundary is missing")
	}
	required := []string{
		"needs: strategy",
		"GOTOOLCHAIN: go1.27.0",
		"persist-credentials: false",
		"go run ./cmd/language-readiness-witness/proposal-promotion",
		`--check "$RUNNER_TEMP/language-readiness-proposal-promotion-a.json"`,
		"name: language-readiness-proposal-promotion-${{ env.HEAD_SHA }}",
		"name: language-readiness-proposal-promotion-v2-${{ env.HEAD_SHA }}",
	}
	for _, fragment := range required {
		if !strings.Contains(producer, fragment) {
			t.Fatalf("promotion evidence workflow is missing %q", fragment)
		}
	}
	for _, forbidden := range []string{
		"needs: proposal-continuity",
		"promotion-authorized-continuity",
		"guarded-promotion-receipt",
	} {
		if strings.Contains(producer, forbidden) {
			t.Fatalf("promotion evidence producer depends on judge %q", forbidden)
		}
	}
}
