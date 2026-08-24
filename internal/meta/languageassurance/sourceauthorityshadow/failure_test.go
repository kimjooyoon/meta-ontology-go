package sourceauthorityshadow

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityeval"
)

func TestObservePreservesUnknownEvidence(t *testing.T) {
	raw, sha := fixture(t)
	bundle, err := sourceauthorityeval.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Authorities = nil
	raw, _ = json.Marshal(bundle)
	report := Observe(raw, sha)
	if report.Observation != "UNKNOWN" || report.Resolution != "INVARIANT_ONLY" ||
		report.Enforcement != "BLOCK" || report.Evaluation.Observation != "UNKNOWN" {
		t.Fatalf("unknown receipt = %#v", report)
	}
}

func TestObserveLowersMismatchedHeadResolution(t *testing.T) {
	raw, _ := fixture(t)
	report := Observe(raw, strings.Repeat("b", 40))
	if report.Observation != "UNKNOWN" || report.Resolution != "INVARIANT_ONLY" ||
		report.Enforcement != "BLOCK" || report.Reason != "SOURCE_AUTHORITY_SHADOW_HEAD_UNKNOWN" {
		t.Fatalf("head mismatch = %#v", report)
	}
}

func TestObservePreservesKnownMismatch(t *testing.T) {
	raw, sha := fixture(t)
	bundle, err := sourceauthorityeval.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Facts[0].Claim = []byte("different")
	bundle.Facts[0].ClaimDigest = sourceauthorityeval.DigestBytes(bundle.Facts[0].Claim)
	raw, _ = json.Marshal(bundle)
	report := Observe(raw, sha)
	if report.Observation != "NOT_SATISFIED" || report.Resolution != "EXACT" ||
		report.Enforcement != "BLOCK" || report.Evaluation.Summary.FailedFacts != 1 {
		t.Fatalf("known mismatch = %#v", report)
	}
}

func TestObserveMalformedInputFailsClosed(t *testing.T) {
	report := Observe([]byte("{"), strings.Repeat("a", 40))
	if report.Observation != "UNKNOWN" || report.Resolution != "INVARIANT_ONLY" ||
		report.Enforcement != "BLOCK" || report.PromotionCreditBPS != 0 {
		t.Fatalf("malformed receipt = %#v", report)
	}
}
