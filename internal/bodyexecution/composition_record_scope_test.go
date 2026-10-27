package bodyexecution

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCompositionRecordScopeKeepsWholeDeclarationOrigin(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/selected-record-scope.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	first, err := GenerateComputedProjection(context.Background(), "scope.gooo", source, "Evaluate")
	if err != nil || len(first.Plan.Records) != 20 {
		t.Fatal("full declaration view lost", len(first.Plan.Records), err)
	}
	extended := append(append([]byte(nil), source...), []byte("\nentity Later id \"urn:scope:later\" fields {\nfield n id \"urn:scope:later:n\" type integer required one\n}\n")...)
	second, err := GenerateComputedProjection(context.Background(), "scope.gooo", extended, "Evaluate")
	if err != nil || len(second.Plan.Records) != 21 {
		t.Fatal("extended declaration view", len(second.Plan.Records), err)
	}
	if first.SourceSHA256 == second.SourceSHA256 || first.Plan.SemanticFingerprint == second.Plan.SemanticFingerprint {
		t.Fatal("unselected source was removed from origin")
	}
}

func TestCompositionRecordScopeChargesTheUnionOfSelectedBindStages(t *testing.T) {
	var source strings.Builder
	source.WriteString("package union\nnamespace union\n")
	for i := range 17 {
		fmt.Fprintf(&source, "entity R%d id \"urn:union:%d\" fields {\nfield value id \"urn:union:%d:value\" type integer required one\n}\n", i, i, i)
	}
	for stage := range 2 {
		name := []string{"First", "Second"}[stage]
		fmt.Fprintf(&source, "activity %s(R0) -> R0 computes `\nlet total = input.value\n", name)
		for i := 1 + stage*8; i <= 8+stage*8; i++ {
			fmt.Fprintf(&source, "let r%d = R%d{value: input.value}\ntotal = total + r%d.value\n", i, i, i)
		}
		source.WriteString("return R0{value: total}\n`\n")
	}
	source.WriteString("bind First.result -> Second.input\n")
	result, err := GenerateComputedProjection(context.Background(), "union.gooo", []byte(source.String()), "Second")
	if result.Source != "" || err == nil || !strings.Contains(err.Error(), "selected activity closure") || !strings.Contains(err.Error(), "supports at most 16") {
		t.Fatal("per-stage counting widened the complete operation", err)
	}
}
