package metarecognition

import (
	"testing"
)

func TestClosedCorpus(t *testing.T) {
	cases := Corpus()
	if len(cases) < 17 {
		t.Fatalf("closed corpus cases = %d, want at least 17", len(cases))
	}
	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			gooo, baseline := Evaluate(c)
			if gooo.State != c.Expected.State || gooo.Reason != c.Expected.Reason || !equalIDs(gooo.LocalizedIDs, c.Expected.LocalizedIDs) {
				t.Fatalf("gooo = %#v, want %#v", gooo, c.Expected)
			}
			if baseline.State != c.Expected.State || baseline.Reason != c.Expected.Reason || !equalIDs(baseline.LocalizedIDs, c.Expected.LocalizedIDs) {
				t.Fatalf("baseline = %#v, want %#v", baseline, c.Expected)
			}
		})
	}
}
func TestCanonicalReplayAndManifest(t *testing.T) {
	cases := Corpus()
	forward, err := Run(cases)
	if err != nil {
		t.Fatal(err)
	}
	reverseCases := append([]Case(nil), cases...)
	for left, right := 0, len(reverseCases)-1; left < right; left, right = left+1, right-1 {
		reverseCases[left], reverseCases[right] = reverseCases[right], reverseCases[left]
	}
	reverse, err := Run(reverseCases)
	if err != nil {
		t.Fatal(err)
	}
	forwardJSON, err := forward.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	reverseJSON, err := reverse.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(forwardJSON) != string(reverseJSON) {
		t.Fatal("canonical manifest changed under case permutation")
	}
	if forward.Finding != NoUniqueBenefit {
		t.Fatalf("finding = %s, want %s", forward.Finding, NoUniqueBenefit)
	}
	if !forward.Summary.ExactOutcomeVector || !forward.Summary.ExactReasonLocalizationVector {
		t.Fatal("comparison vectors are not exact")
	}
	if forward.Summary.GoooFalsePasses != 0 || forward.Summary.GoooFalseNegatives != 0 || forward.Summary.BaselineFalsePasses != 0 || forward.Summary.BaselineFalseNegatives != 0 {
		t.Fatalf("false-pass/negative counts = %#v", forward.Summary)
	}
	if forward.Summary.BaselineWorkUnits > forward.Summary.GoooWorkUnits || !forward.Summary.GoooRatio.Known || !forward.Summary.BaselineRatio.Known {
		t.Fatalf("fair-baseline work/ratio rule failed: %#v", forward.Summary)
	}
	digest, err := ManifestDigest(forward)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("manifest_json=%s", forwardJSON)
	t.Logf("manifest_digest=%s", digest)
}
