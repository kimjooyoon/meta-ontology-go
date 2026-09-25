package sourceauthorityeval

import "testing"

func TestKnownViolationsBlockExactly(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Bundle)
		reason string
	}{
		{"missing source binding",
			func(b *Bundle) { b.Facts[0].SourceRef = "" },
			"REQUIRED_BINDING_MISSING"},
		{"snapshot mismatch",
			func(b *Bundle) { b.Facts[0].SourceSnapshotDigest = DigestBytes([]byte("x")) },
			"SOURCE_SNAPSHOT_MISMATCH"},
		{"invalid span",
			func(b *Bundle) { b.Facts[0].Span.End++ },
			"SOURCE_SPAN_INVALID"},
		{"span digest mismatch",
			func(b *Bundle) { b.Facts[0].Span.Digest = DigestBytes([]byte("x")) },
			"SOURCE_SPAN_DIGEST_MISMATCH"},
		{"missing authority binding",
			func(b *Bundle) { b.Facts[0].AuthorityRef = "" },
			"REQUIRED_BINDING_MISSING"},
		{"authority scope mismatch",
			func(b *Bundle) { b.Authorities[0].End-- },
			"AUTHORITY_SCOPE_MISMATCH"},
		{"semantic interpretation",
			func(b *Bundle) {
				b.Facts[0].Claim = []byte("interpretation")
				b.Facts[0].ClaimDigest = DigestBytes(b.Facts[0].Claim)
			},
			"CLAIM_NOT_EXACT_SOURCE_BYTES"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			bundle := exactBundle()
			test.mutate(&bundle)
			report := Evaluate(bundle)
			if report.Observation != "NOT_SATISFIED" ||
				report.Resolution != "EXACT" || report.Enforcement != "BLOCK" {
				t.Fatalf("outcome = %s/%s/%s", report.Observation, report.Resolution, report.Enforcement)
			}
			if report.Receipts[0].Reason != test.reason {
				t.Fatalf("reason = %q want %q", report.Receipts[0].Reason, test.reason)
			}
		})
	}
}

func TestUnknownAndErrorRemainDistinct(t *testing.T) {
	missing := exactBundle()
	missing.Sources = nil
	unknown := Evaluate(missing)
	if unknown.Observation != "UNKNOWN" || unknown.Resolution != "INVARIANT_ONLY" || unknown.Enforcement != "BLOCK" {
		t.Fatalf("unknown outcome = %+v", unknown)
	}
	duplicate := exactBundle()
	duplicate.Sources = append(duplicate.Sources, duplicate.Sources[0])
	failed := Evaluate(duplicate)
	if failed.Observation != "ERROR" || failed.Resolution != "EXACT" || failed.Enforcement != "BLOCK" {
		t.Fatalf("error outcome = %+v", failed)
	}
}

func TestEmptyAcceptedDenominatorIsUnknown(t *testing.T) {
	bundle := exactBundle()
	bundle.Facts[0].State = "CANDIDATE"
	report := Evaluate(bundle)
	if report.Reason != "SOURCE_AUTHORITY_DENOMINATOR_EMPTY" ||
		report.Observation != "UNKNOWN" || report.Enforcement != "BLOCK" {
		t.Fatalf("empty denominator outcome = %+v", report)
	}
}
