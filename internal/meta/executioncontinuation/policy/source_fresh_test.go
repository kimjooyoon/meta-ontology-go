package policy

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"os"
	"testing"
)

func TestConsumerPolicyProjectionFresh1617(t *testing.T) {
	for _, c := range []struct{ activity, file string }{{"BindSubjects", "subjects.generated.go"}, {"PermitContinuation", "eligibility.generated.go"}, {"PermitWait", "wait.generated.go"}, {"BindTargetRun", "target.generated.go"}, {"ObserveStage", "stage.generated.go"}, {"ArtifactLabel", "artifact.generated.go"}, {"WorkflowPath", "workflow.generated.go"}} {
		t.Run(c.activity, func(t *testing.T) {
			members, err := policySources()
			if err != nil {
				t.Fatal(err)
			}
			p, err := bodycodegen.GenerateSourceBundle(members, c.activity)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(c.file)
			if err != nil || string(raw) != p.Source {
				t.Fatal("projection differs", err)
			}
		})
	}
}

func TestConsumerPolicyDecisionsFresh1617(t *testing.T) {
	cases := []struct{ declared, grant, contract, want string }{{"s", "s", "s", "SUBJECT_BOUND"}, {"", "s", "s", "SUBJECT_UNKNOWN"}, {"", "s", "t", "SUBJECT_REFUTED"}, {"s", "", "t", "SUBJECT_REFUTED"}}
	for _, c := range cases {
		if got := BindSubjects(c.declared, c.grant, c.contract); got != c.want {
			t.Fatal(c, got)
		}
	}
	if PermitWait("COMPLETED", true) || !PermitWait("CREATED", true) || !PermitWait("NONE", true) {
		t.Fatal("wait cycle")
	}
	if BindTargetRun(1, 1, 1) != "TARGET_RUN_REFUTED" || BindTargetRun(1, 2, 0) != "TARGET_RUN_UNKNOWN" || BindTargetRun(1, 2, 3) != "TARGET_RUN_REFUTED" || BindTargetRun(1, 2, 2) != "TARGET_RUN_BOUND" {
		t.Fatal("run identity")
	}
	for _, c := range []struct {
		sent, observed, executed bool
		want                     string
	}{{false, false, false, "DISPATCH_UNKNOWN"}, {true, false, false, "DISPATCH_SENT"}, {true, true, false, "TARGET_RUN_OBSERVED"}, {true, true, true, "EXECUTION_OBSERVED"}, {true, false, true, "CONTINUATION_REFUTED"}} {
		if ObserveStage(c.sent, c.observed, c.executed) != c.want {
			t.Fatal(c)
		}
	}
	if !PermitContinuation("CLOSED", "GRANTED_UNCONSUMED", 1, 0, 0, true) || PermitContinuation("CLOSED", "GRANTED_UNCONSUMED", 0, 1, 0, true) || PermitContinuation("UNKNOWN", "GRANTED_UNCONSUMED", 1, 0, 0, true) {
		t.Fatal("grant boundary")
	}
}
