package sourceauthorityshadow

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

const fixturePath = "../../../../examples/language-assurance-kernel/source-authority-shadow.json"

func fixture(t *testing.T) ([]byte, string) {
	t.Helper()
	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.Repeat("a", 40)
	raw = bytes.Replace(raw, []byte(strings.Repeat("0", 40)), []byte(sha), 1)
	return raw, sha
}

func TestObserveRecordsPinnedAuthorityWithoutPromotion(t *testing.T) {
	raw, sha := fixture(t)
	report, replay := Observe(raw, sha), Observe(raw, sha)
	if report.Schema != ReceiptSchema || report.Mode != Mode || report.SubjectSHA != sha ||
		report.Observation != "SATISFIED" || report.Resolution != "EXACT" ||
		report.Enforcement != "ALLOW" || report.GateEffect != "NO_EFFECT" {
		t.Fatalf("shadow receipt = %#v", report)
	}
	if report.PromotionCreditBPS != 0 || report.RepositoryWrites != 0 ||
		report.Evaluation.Summary.AcceptedFacts != 1 || report.Evaluation.Summary.BackedFacts != 1 ||
		report.Evaluation.Summary.CoverageBPS != 10000 || len(report.Evaluation.Receipts) != 1 {
		t.Fatalf("shadow summary = %#v", report)
	}
	if len(report.Indicators) != 6 || report.ReceiptDigest == "" ||
		report.ReceiptDigest != replay.ReceiptDigest {
		t.Fatalf("shadow replay = %#v", report)
	}
	counts := map[string]int{}
	for _, indicator := range report.Indicators {
		counts[indicator.Class]++
	}
	if counts["OUTCOME"] != 1 || counts["DRIVER"] != 2 || counts["GUARDRAIL"] != 3 {
		t.Fatalf("indicator classes = %#v", counts)
	}
}
