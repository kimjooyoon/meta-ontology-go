package feedbackstate

import (
	"encoding/json"
	"os"
	"testing"
)

type semanticUseCase struct {
	ID, Decision, SourceDecision, FromResolution, ToResolution string
	Mutation, WantDecision, WantReason                          string
	PreviousDescents, Descents                                 int
}

func TestExecutableSemanticCycleUseCases(t *testing.T) {
	data, err := os.ReadFile("../../../examples/feedback-semantic-cycle/usecases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []semanticUseCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 9 {
		t.Fatalf("use cases = %d", len(cases))
	}
	for _, test := range cases {
		t.Run(test.ID, func(t *testing.T) {
			input := fixture(test.Decision, test.SourceDecision, test.FromResolution,
				test.ToResolution, test.PreviousDescents, test.Descents)
			switch test.Mutation {
			case "payload-digest":
				input.PayloadDigest = "sha256:wrong"
			case "repository-write":
				input.RepositoryWrites = 1
			}
			report := Evaluate(input)
			if report.Decision != test.WantDecision || report.Reason != test.WantReason {
				t.Fatalf("got %s/%s", report.Decision, report.Reason)
			}
		})
	}
}
