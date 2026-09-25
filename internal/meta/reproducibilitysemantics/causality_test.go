package reproducibilitysemantics_test

import (
	"strings"
	"testing"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/reproducibilitysemantics"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/reproducibilitysemanticsconsumer"
)

func TestSemanticAndPresentationInterventionsAreSeparate(t *testing.T) {
	head := strings.Repeat("d", 40)
	baseReceipt, base := judgedFixture(t, head, []byte(fixtureSource))
	semanticSource := strings.Replace(fixtureSource, "meaning.observed=meaning/render-approved/v1",
		"meaning.observed=meaning/charge-and-ledger/v1", 1)
	semanticReceipt, semantic := judgedFixture(t, head, []byte(semanticSource))
	formattedSource := "// presentation-only comment\n" + strings.ReplaceAll(fixtureSource, "\n", "\n ")
	formattedReceipt, formatted := judgedFixture(t, head, []byte(formattedSource))
	if semantic.Summary.MeaningClaim == base.Summary.MeaningClaim || semantic.Summary.JointClaim == base.Summary.JointClaim {
		t.Fatalf("semantic intervention did not change coordinates: base=%#v semantic=%#v", base.Summary, semantic.Summary)
	}
	if semanticReceipt.SemanticDigest == baseReceipt.SemanticDigest || semanticReceipt.SourceDigest == baseReceipt.SourceDigest {
		t.Fatal("semantic intervention did not change both source and semantic digests")
	}
	if formattedReceipt.SourceDigest == baseReceipt.SourceDigest || formattedReceipt.SemanticDigest != baseReceipt.SemanticDigest ||
		formatted.Summary != base.Summary {
		t.Fatalf("presentation intervention changed semantic result unexpectedly: base=%#v formatted=%#v", base.Summary, formatted.Summary)
	}
	artifact, err := consumer.BuildInterventionArtifact(base, semantic, formatted)
	if err != nil {
		t.Fatal(err)
	}
	if err := consumer.ValidateIntervention(artifact); err != nil {
		t.Fatal(err)
	}
	if artifact.Denominator != 2 || len(artifact.Cases) != 2 || artifact.Cases[0].TransitionsBefore[1].MeaningTransition == artifact.Cases[0].TransitionsAfter[1].MeaningTransition || artifact.Cases[1].TransitionsBefore[1].MeaningTransition != artifact.Cases[1].TransitionsAfter[1].MeaningTransition {
		t.Fatalf("intervention artifact = %#v", artifact)
	}
}

func judgedFixture(t *testing.T, head string, source []byte) (producer.Receipt, consumer.Judgment) {
	t.Helper()
	receipt := producer.Produce("fixture.gooo", head, source)
	raw := receiptJSON(t, receipt)
	judgment := consumer.Judge("fixture.gooo", head, source, raw)
	if err := consumer.ValidateJudgment("fixture.gooo", head, source, raw, judgment); err != nil {
		t.Fatal(err)
	}
	return receipt, judgment
}

func TestDigestOnlyReceiptIsRefuted(t *testing.T) {
	head := strings.Repeat("e", 40)
	source := []byte(fixtureSource)
	receipt := producer.Produce("fixture.gooo", head, source)
	receipt.SemanticDigest = ""
	judgment := consumer.Judge("fixture.gooo", head, source, receiptJSON(t, receipt))
	if judgment.Decision != consumer.StatusRefuted || judgment.Reason != "DIGEST_ONLY_REFUTED" {
		t.Fatalf("judgment = %#v", judgment)
	}
}
