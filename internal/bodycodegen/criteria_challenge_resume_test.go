package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestCriteriaChallengeResumeRechecksWithoutRewritingHistory(t *testing.T) {
	source := criteriaChallengeFixture()
	prior := generateCriteria(t, source)
	changed := strings.Replace(source, "return input1.value == 0-input0", "return input1.value == input0", 1)
	r, err := ResumeRecordAssembly(context.Background(), "challenge.gooo", []byte(source), []byte(changed), prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	a := r.Report.RecordAssembly
	if a.Status != "COMPLETE_FINITE" || a.Criteria.Passed != 2 || a.Continuation.RecheckedAttempts != 2 || a.Continuation.AddedAttempts != 0 ||
		!equalLocalCallEvidence(a.CriteriaQuestions, prior.Report.RecordAssembly.CriteriaQuestions) {
		t.Fatal("predicate edit rewrote earlier challenge", a)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "challenge.gooo", []byte(changed), r); err != nil {
		t.Fatal(err)
	}
}
