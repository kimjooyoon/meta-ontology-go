package consumer

import (
	"strings"
	"testing"
	"testing/fstest"
)

const semanticSource = `package hygienicoriginidentity
namespace hygienicoriginidentity
entity OriginIdentity id "gooo://hygienic-origin-identity/entity/origin-identity"
entity ScopeProvenance id "gooo://hygienic-origin-identity/entity/scope-provenance"
entity GeneratedName id "gooo://hygienic-origin-identity/entity/generated-name"
entity ConsumerBinding id "gooo://hygienic-origin-identity/entity/consumer-binding"
entity CapturedResult id "gooo://hygienic-origin-identity/entity/captured-result"
entity HygienicResult id "gooo://hygienic-origin-identity/entity/hygienic-result"
entity CapturedOriginClaim id "gooo://hygienic-origin-identity/entity/captured-origin-claim"
entity CapturedScopeClaim id "gooo://hygienic-origin-identity/entity/captured-scope-claim"
entity HygienicOriginClaim id "gooo://hygienic-origin-identity/entity/hygienic-origin-claim"
entity HygienicScopeClaim id "gooo://hygienic-origin-identity/entity/hygienic-scope-claim"
entity ProofReceipt id "gooo://hygienic-origin-identity/entity/proof-receipt"
activity ProduceCapturedName(OriginIdentity) -> GeneratedName computes "hoi.produce case=captured spelling=tmp origin=consumer-binding definition-scope=consumer-call-site use-scope=consumer-call-site"
activity ProduceHygienicName(OriginIdentity) -> GeneratedName computes "hoi.produce case=hygienic spelling=tmp origin=producer-expansion-1 definition-scope=fresh-producer-expansion-1 use-scope=fresh-producer-expansion-1"
activity ConsumeCapturedName(GeneratedName) -> ConsumerBinding computes "hoi.resolve case=captured binding=consumer-binding use-scope=consumer-call-site"
activity ConsumeHygienicName(GeneratedName) -> ConsumerBinding computes "hoi.resolve case=hygienic binding=producer-expansion-1 use-scope=fresh-producer-expansion-1"
activity ObserveCapturedResult(ConsumerBinding) -> CapturedResult
activity ObserveHygienicResult(ConsumerBinding) -> HygienicResult
activity PreserveOriginIdentity(GeneratedName) -> OriginIdentity
activity PreserveScopeProvenance(GeneratedName) -> ScopeProvenance
activity EmitProofReceipt(CapturedResult) -> ProofReceipt
# non-authoritative comment
`

func TestCommentOnlyChangePreservesSemanticReport(t *testing.T) {
	files := fstest.MapFS{
		"main.gooo":    {Data: []byte(semanticSource)},
		"comment.gooo": {Data: []byte(strings.Replace(semanticSource, "# non-authoritative comment", "# another comment only", 1))},
	}
	mainReport, err := Evaluate(files, "main.gooo", strings.Repeat("a", 40), SnapshotPair{})
	if err != nil {
		t.Fatal(err)
	}
	commentReport, err := Evaluate(files, "comment.gooo", strings.Repeat("a", 40), SnapshotPair{})
	if err != nil {
		t.Fatal(err)
	}
	if mainReport.Source.RawDigest == commentReport.Source.RawDigest || mainReport.Source.SemanticDigest != commentReport.Source.SemanticDigest || ContentDigest(mainReport) != ContentDigest(commentReport) {
		t.Fatalf("comment-only change was semantic: main=%#v comment=%#v", mainReport.Source, commentReport.Source)
	}
}

func TestUnknownComesFromSemanticValue(t *testing.T) {
	source := strings.Replace(semanticSource, "hoi.resolve case=hygienic binding=producer-expansion-1 use-scope=fresh-producer-expansion-1", "hoi.resolve case=hygienic provenance=missing", 1) + "# misleading stage=scope-resolution step=ignored reason=ignored\n"
	report, err := Evaluate(fstest.MapFS{"unknown.gooo": {Data: []byte(source)}}, "unknown.gooo", strings.Repeat("b", 40), SnapshotPair{})
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionUnknown || len(report.Unknowns) != 1 || report.Unknowns[0].EvidenceDigest == "" || report.Unknowns[0].Provenance == "" {
		t.Fatalf("unknown was not semantic: %#v", report)
	}
	if report.Unknowns[0].Stage != "scope-resolution" || report.Unknowns[0].Step != "resolve-generated-binding" || report.Unknowns[0].Reason != "scope-provenance-absent" {
		t.Fatalf("unknown coordinates changed: %#v", report.Unknowns[0])
	}
}
