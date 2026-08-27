package verify

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/hygienicoriginidentity/consumer"
)

func TestValidateRejectsCoherentResealedTamper(t *testing.T) {
	source := `package hygienicoriginidentity
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
activity ProduceCapturedName(OriginIdentity) -> GeneratedName computes "hoi.produce case=captured spelling=tmp origin=consumer-binding scope=consumer-call-site"
activity ProduceHygienicName(OriginIdentity) -> GeneratedName computes "hoi.produce case=hygienic spelling=tmp origin=producer-expansion-1 scope=fresh-producer-expansion-1"
activity ConsumeCapturedName(GeneratedName) -> ConsumerBinding computes "hoi.resolve case=captured binding=consumer-binding"
activity ConsumeHygienicName(GeneratedName) -> ConsumerBinding computes "hoi.resolve case=hygienic binding=producer-expansion-1"
activity ObserveCapturedResult(ConsumerBinding) -> CapturedResult
activity ObserveHygienicResult(ConsumerBinding) -> HygienicResult
activity PreserveOriginIdentity(GeneratedName) -> OriginIdentity
activity PreserveScopeProvenance(GeneratedName) -> ScopeProvenance
activity EmitProofReceipt(CapturedResult) -> ProofReceipt
`
	files := fstest.MapFS{"main.gooo": {Data: []byte(source)}}
	report, err := consumer.Evaluate(files, "main.gooo", strings.Repeat("a", 40), consumer.SnapshotPair{})
	if err != nil {
		t.Fatal(err)
	}
	report.Cases[1].ResolvedIdentity = consumer.ConsumerBinding
	report = consumer.Seal(report)
	if err := Validate(files, report, ExpectationPass, strings.Repeat("a", 40)); err == nil {
		t.Fatal("coherent resealed tamper was accepted")
	}
}
