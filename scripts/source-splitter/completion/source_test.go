package completion

import (
	"bytes"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestCompletionPolicyIsTheCurrentGoooProjection(t *testing.T) {
	raw, err := os.ReadFile("source.gooo")
	if err != nil || !bytes.Equal(raw, source) {
		t.Fatal("compiled policy differs from the current authored source", err)
	}
	for activity, file := range map[string]string{"Classify": "classify.generated.go", "Decision": "decision.generated.go",
		"Resolution": "resolution.generated.go", "NextOperation": "next.generated.go"} {
		generated, err := bodycodegen.Generate("source.gooo", raw, activity)
		if err != nil {
			t.Fatal(err)
		}
		current, err := os.ReadFile(file)
		if err != nil || !bytes.Equal(current, []byte(generated.Source)) {
			t.Fatal("regenerate the source-owned completion policy", activity, err)
		}
	}
}

func TestCompletionPolicyRetainsIndependentOutcomes(t *testing.T) {
	for _, row := range completionCases {
		t.Run(row.name, func(t *testing.T) {
			reason := Classify(row.counts[0], row.counts[1], row.counts[2], row.counts[3])
			if reason != row.reason || Decision(reason) != row.decision ||
				Resolution(reason) != row.resolution || NextOperation(reason) != row.next {
				t.Fatal("unexpected completion tuple", reason, Decision(reason), Resolution(reason), NextOperation(reason))
			}
		})
	}
	if Decision("future-observation") != "UNKNOWN" || Resolution("future-observation") != "UNKNOWN" ||
		NextOperation("future-observation") != "restore-split-observation" {
		t.Fatal("unsupported observations were treated as completion")
	}
}
