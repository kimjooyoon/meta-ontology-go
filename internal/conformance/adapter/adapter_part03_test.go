package adapter

import (
	"bytes"
	"strings"
	"testing"
)

func TestDeferredOracleIsVisibleAndNotPromotable(t *testing.T) {
	request := sampleRequest(StatusDeferred)
	response := sampleResponse(StatusDeferred, false)
	evaluation := Evaluate(request, response)
	if !evaluation.Matched || evaluation.ExitCode != ExitOK {
		t.Fatalf("expected deferred result was rejected: %+v", evaluation)
	}
	request.Expected.Status = StatusPass
	evaluation = Evaluate(request, response)
	if evaluation.Matched || evaluation.ExitCode != ExitDeferred {
		t.Fatalf("deferred result was hidden as a normal mismatch: %+v", evaluation)
	}
	if response.PromotionEligible {
		t.Fatal("deferred response was promotion eligible")
	}
}
func TestEvidenceProjectionIgnoresProducerAndFactOrder(t *testing.T) {
	response := sampleResponse(StatusPass, false)
	goEvidence, err := response.ProjectEvidence("go", StageGoBaseline)
	if err != nil {
		t.Fatal(err)
	}
	goooEvidence, err := response.ProjectEvidence("gooo", StageGoBaseline)
	if err != nil {
		t.Fatal(err)
	}
	goooEvidence.Bundle.Facts = append([]EvidenceFact(nil), reverseFacts(goooEvidence.Bundle.Facts)...)
	if err := CompareEvidence(goEvidence, goooEvidence); err != nil {
		t.Fatal(err)
	}
	manifest, err := goEvidence.ManifestJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(manifest, []byte{'\n'}) || !strings.Contains(string(manifest), `"producer":"go"`) {
		t.Fatalf("manifest lost deterministic identity: %s", manifest)
	}
	response.Observed.SemanticDigest = "sha256:different"
	different, err := response.ProjectEvidence("gooo", StageGoBaseline)
	if err != nil {
		t.Fatal(err)
	}
	if err := CompareEvidence(goEvidence, different); err == nil {
		t.Fatal("different evidence was accepted")
	}
}
func sampleRequest(status Status) Request {
	return Request{
		Schema: ProtocolSchema, Fixture: "billing/main", Operation: OperationGenerate, RunID: "run-001",
		Input:    Input{DSL: "entity billing", IR: []byte(`{"type":"entity"}`), SourceURI: "examples/billing/main.gooo"},
		Contract: Contract{AST: "ast/v1", IR: "ir/v1", Generator: "generator/v1", Marker: "marker/v1", PolicyHash: "policy"},
		Options:  Options{CanonicalOutput: true}, Expected: Expectation{Status: status},
	}
}
