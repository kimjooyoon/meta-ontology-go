package languagesourceexecution

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

const source = `package billing
namespace billing
entity In id "billing://entity/in"
entity Out id "billing://entity/out"
activity Execute(In) -> Out
`

func testInput(t *testing.T) Input {
	t.Helper()
	marshal := func(receipt sourceexecution.Receipt) []byte {
		raw, err := sourceexecution.Marshal(receipt)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	positive := marshal(sourceexecution.Execute(sourceexecution.Request{"fixture.gooo", source, "Execute"}))
	return Input{Contract: CanonicalContract(), HeadSHA: "0123456789012345678901234567890123456789",
		Positive: positive, Replay: append([]byte(nil), positive...),
		UnknownEntry: marshal(sourceexecution.Execute(sourceexecution.Request{"fixture.gooo", source, "Missing"})),
		InvalidSyntax: marshal(sourceexecution.Execute(sourceexecution.Request{"broken.gooo", "activity", "Missing"}))}
}

func TestEvaluateSourceExecutionContract(t *testing.T) {
	artifact := Evaluate(testInput(t))
	if artifact.Decision != "PASS" || artifact.Summary.CasesSatisfied != 4 ||
		artifact.Summary.SourceExecutions != 1 || artifact.Summary.DeterministicReplays != 1 ||
		artifact.Summary.DiagnosticRejections != 2 || artifact.Digest == "" {
		t.Fatalf("artifact=%#v", artifact)
	}
}

func TestUnknownTopDecisionLowersResolution(t *testing.T) {
	input := testInput(t)
	var receipt sourceexecution.Receipt
	if err := json.Unmarshal(input.Positive, &receipt); err != nil {
		t.Fatal(err)
	}
	receipt.Decision = "UNKNOWN"
	input.Positive, _ = json.Marshal(receipt)
	artifact := Evaluate(input)
	if artifact.Decision != "FAIL_CLOSED" || artifact.Resolution != "LOWER_RESOLUTION" || artifact.Summary.Unknowns == 0 {
		t.Fatalf("artifact=%#v", artifact)
	}
}
