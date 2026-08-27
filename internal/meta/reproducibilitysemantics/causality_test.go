package reproducibilitysemantics

import (
	"strings"
	"testing"
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
}

func judgedFixture(t *testing.T, head string, source []byte) (Receipt, Judgment) {
	t.Helper()
	receipt := Produce("fixture.gooo", head, source)
	judgment := Judge("fixture.gooo", head, source, receipt)
	if err := ValidateJudgment("fixture.gooo", head, source, receipt, judgment); err != nil {
		t.Fatal(err)
	}
	return receipt, judgment
}

func TestDigestOnlyReceiptIsRefuted(t *testing.T) {
	head := strings.Repeat("e", 40)
	source := []byte(fixtureSource)
	receipt := Produce("fixture.gooo", head, source)
	receipt.SemanticDigest = ""
	judgment := Judge("fixture.gooo", head, source, receipt)
	if judgment.Decision != StatusRefuted || judgment.Reason != "DIGEST_ONLY_REFUTED" {
		t.Fatalf("judgment = %#v", judgment)
	}
}
