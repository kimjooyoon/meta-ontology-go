package verify

import (
	"strings"
	"testing"

	termination "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtermination"
)

func TestIndependentJudgeRecomputesCycleInsteadOfTrustingReceipt(t *testing.T) {
	input := termination.Input{Schema: termination.InputSchema, Repository: "kimjooyoon/meta-ontology-go",
		Subject: termination.Consumer, Producer: termination.Producer, Consumer: termination.Consumer,
		MetaOperation: termination.MetaOperation, ProofChoice: termination.ProofChoice, Stage: termination.TraceStage,
		MaxSteps: 4, Trace: []termination.Observation{
			{Stage: termination.TraceStage, Step: 1, BeforeState: digest("a"), AfterState: digest("b"),
				BeforeRank: 1, AfterRank: 2, Decision: "CHANGED", Reason: "METAPROGRAM_STATE_CHANGED"},
			{Stage: termination.TraceStage, Step: 2, BeforeState: digest("b"), AfterState: digest("a"),
				BeforeRank: 2, AfterRank: 1, Decision: "CHANGED", Reason: "METAPROGRAM_STATE_CHANGED"},
		}}
	receipt, err := termination.Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Decision = termination.DecisionFixedPoint
	if _, err := Verify(input, receipt); err == nil {
		t.Fatal("judge trusted a forged fixed-point decision")
	}
}

func digest(value string) string { return "sha256:" + strings.Repeat(value, 64) }
