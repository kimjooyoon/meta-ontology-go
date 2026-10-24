package generation

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestExtractionPrerequisiteRejectsSubstitutedReadiness(t *testing.T) {
	head := strings.Repeat("a", 40)
	binding, _ := BindingForOperation(DefaultRegistry(), sourcepolicy.OperationExtractFunction)
	action := actionFor(candidate{indicator: metric("x.go:2:Value", sourcepolicy.OperationExtractFunction, false, false),
		binding: binding}, "sha256:"+strings.Repeat("b", 64))
	ready := NewCandidatePrerequisite(action, head, strings.Repeat("c", 64), strings.Repeat("d", 64), "EXTRACTION_PLAN_READY")
	ready.Stage, ready.Step = "prepare-extraction", "stage-without-writing"
	if !validCandidatePrerequisite(ready, action, head) {
		t.Fatal("valid observation rejected")
	}
	for name, mutate := range map[string]func(*CandidatePrerequisite){
		"stage":       func(p *CandidatePrerequisite) { p.Stage = "" },
		"source":      func(p *CandidatePrerequisite) { p.SourceDigest = "" },
		"uncertainty": func(p *CandidatePrerequisite) { p.UnknownClass = "DIRECT_MISSING" },
		"diagnostic":  func(p *CandidatePrerequisite) { p.Diagnostics = []string{"CALLEE_EFFECTS_UNPROVEN"} },
		"state":       func(p *CandidatePrerequisite) { p.State = "REFUTED"; p.Reason = "CALLEE_EFFECTS_UNPROVEN" },
	} {
		t.Run(name, func(t *testing.T) {
			forged := ready
			mutate(&forged)
			if validCandidatePrerequisite(forged, action, head) {
				t.Fatal("invalid readiness accepted", forged)
			}
		})
	}
}
