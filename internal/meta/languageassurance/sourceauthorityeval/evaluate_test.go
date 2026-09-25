package sourceauthorityeval

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthority"
)

func exactBundle() Bundle {
	content := []byte("authoritative fact")
	digest := DigestBytes(content)
	return Bundle{
		Schema:         InputSchema,
		SubjectSHA:     strings.Repeat("a", 40),
		ContractDigest: sourceauthority.Digest(),
		Sources: []Source{{
			ID: "source-1", URI: "repo://fixture/source.gooo",
			SnapshotDigest: digest, Bytes: content,
		}},
		Authorities: []Authority{{
			ID: "authority-1", SourceRef: "source-1",
			SnapshotDigest: digest, Start: 0, End: len(content),
		}},
		Facts: []Fact{{
			ID: "fact-1", State: "ACCEPTED", Claim: content,
			ClaimDigest: digest, SourceRef: "source-1",
			SourceSnapshotDigest: digest,
			Span:                 Span{Start: 0, End: len(content), Digest: digest},
			AuthorityRef:         "authority-1",
		}},
	}
}

func TestEvaluateExactSourceAuthority(t *testing.T) {
	report := Evaluate(exactBundle())
	if report.Observation != "SATISFIED" ||
		report.Resolution != "EXACT" || report.Enforcement != "ALLOW" {
		t.Fatalf("outcome = %s/%s/%s",
			report.Observation, report.Resolution, report.Enforcement)
	}
	if report.Summary.AcceptedFacts != 1 || report.Summary.BackedFacts != 1 ||
		report.Summary.CoverageBPS != 10000 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if report.MetaOperation != sourceauthority.MetaOperation ||
		report.ReceiptDigest == "" {
		t.Fatalf("meta binding or digest missing: %+v", report)
	}
}

func TestObserveIsOrderStableAndStrict(t *testing.T) {
	bundle := exactBundle()
	candidate := bundle.Facts[0]
	candidate.ID, candidate.State = "candidate-1", "CANDIDATE"
	bundle.Facts = append(bundle.Facts, candidate)
	first, _ := json.Marshal(bundle)
	bundle.Facts[0], bundle.Facts[1] = bundle.Facts[1], bundle.Facts[0]
	second, _ := json.Marshal(bundle)
	if Observe(first).ReceiptDigest != Observe(second).ReceiptDigest {
		t.Fatal("input order changed the receipt digest")
	}
	unknown := []byte(`{"schema":"gooo/source-backed-authority-evidence/v1","extra":true}`)
	report := Observe(unknown)
	if report.Observation != "UNKNOWN" || report.Enforcement != "BLOCK" {
		t.Fatalf("strict decode outcome = %+v", report)
	}
}
