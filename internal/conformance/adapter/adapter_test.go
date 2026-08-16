package adapter

import (
	"bytes"
	"strings"
	"testing"
)

func TestRequestCanonicalizationIsOrderIndependent(t *testing.T) {
	left := sampleRequest(StatusPass)
	right := sampleRequest(StatusPass)
	left.Input.IR = []byte(`{"z":2,"a":{"y":4,"x":3}}`)
	right.Input.IR = []byte(`{"a":{"x":3,"y":4},"z":2}`)
	leftPayload, err := left.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	rightPayload, err := right.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(leftPayload, rightPayload) {
		t.Fatalf("equivalent IR was not canonicalized:\n%s\n%s", leftPayload, rightPayload)
	}
	leftDigest, err := left.Digest()
	if err != nil {
		t.Fatal(err)
	}
	rightDigest, err := right.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if leftDigest != rightDigest || len(leftDigest) != 64 {
		t.Fatalf("canonical digest mismatch: %q != %q", leftDigest, rightDigest)
	}
	left.Input.IR = []byte(`{"a":1}{"b":2}`)
	if _, err := left.CanonicalJSON(); err == nil {
		t.Fatal("trailing JSON was accepted")
	}
}

func TestResponseCanonicalizationSortsProtocolSets(t *testing.T) {
	left := sampleResponse(StatusPass, false)
	right := sampleResponse(StatusPass, true)
	leftPayload, err := left.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	rightPayload, err := right.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(leftPayload, rightPayload) {
		t.Fatalf("equivalent response was not canonicalized:\n%s\n%s", leftPayload, rightPayload)
	}
	if got := string(leftPayload); !strings.HasSuffix(got, "\n") {
		t.Fatalf("canonical response is not JSONL: %q", got)
	}
}

func TestCanonicalValidatorRejectsProtectedBoundaryCollisions(t *testing.T) {
	response := sampleResponse(StatusPass, false)
	response.Observed.Regions = append(response.Observed.Regions, Region{
		Kind:       "generated",
		SemanticID: "billing.total",
		Start:      40,
		End:        41,
	})
	if _, err := response.CanonicalJSON(); err == nil || !strings.Contains(err.Error(), "duplicate region") {
		t.Fatalf("duplicate protected boundary was accepted: %v", err)
	}
	response = sampleResponse(StatusPass, false)
	response.Observed.Regions[0].Start = 10
	response.Observed.Regions[0].End = 9
	if _, err := response.CanonicalJSON(); err == nil || !strings.Contains(err.Error(), "invalid range") {
		t.Fatalf("inverted range was accepted: %v", err)
	}
	response = sampleResponse(StatusPass, false)
	response.Observed.Delta = &Delta{Added: []Fact{{
		SubjectID: "billing.total", Predicate: "prov:wasDerivedFrom", ObjectID: "source",
		Class: "prov:Entity", SourceURI: "../outside.gooo", Start: 1, End: 2,
	}}}
	if _, err := response.CanonicalJSON(); err == nil || !strings.Contains(err.Error(), "escape") {
		t.Fatalf("escaping source URI was accepted: %v", err)
	}
}

func TestOracleNW001NegativeOracleRequiresIndependentObserver(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	claim := true
	response.ProducerClaims.NoWrite = &claim
	evaluation := Evaluate(request, response)
	if evaluation.Matched || evaluation.OracleCode != OracleNW001 || evaluation.PromotionEligible {
		t.Fatalf("producer-only negative result was accepted: %+v", evaluation)
	}
	evaluation = EvaluateObserved(request, response, nil)
	if evaluation.Matched || evaluation.OracleCode != OracleNW001 || evaluation.PromotionEligible {
		t.Fatalf("forged producer claim reached observed evaluator: %+v", evaluation)
	}
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	evaluation = EvaluateObserved(request, response, &observation)
	if !evaluation.Matched || evaluation.ExitCode != ExitOK || evaluation.FailureCode != "marker-overlap" || evaluation.PromotionEligible {
		t.Fatalf("valid independent negative result was rejected: %+v", evaluation)
	}
}

func TestRequestValidationRequiresRunBinding(t *testing.T) {
	request := sampleRequest(StatusPass)
	request.RunID = ""
	if err := request.Validate(); err == nil || !strings.Contains(err.Error(), "run_id") {
		t.Fatalf("blank run binding was accepted: %v", err)
	}
}

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

func sampleResponse(status Status, reverse bool) Response {
	regions := []Region{{Kind: "generated", SemanticID: "billing.total", Start: 20, End: 30}, {Kind: "generated", SemanticID: "billing.id", Start: 1, End: 10}}
	slots := []Slot{{ID: "slot.total", OwnerID: "billing.total", Start: 22, End: 28}}
	imports := []Import{{Path: "fmt", UsedBy: []string{"billing.total", "billing.id"}}, {Path: "strings", Alias: "str", UsedBy: []string{"billing.id"}}}
	mappings := []Mapping{{SemanticID: "billing.total", Kind: "field", Source: Range{1, 2}, Generated: Range{20, 30}}}
	if reverse {
		regions[0], regions[1] = regions[1], regions[0]
		imports[0].UsedBy = []string{"billing.id", "billing.total"}
		imports[0], imports[1] = imports[1], imports[0]
	}
	response := Response{
		Schema: ProtocolSchema, Fixture: "billing/main", Operation: OperationGenerate, RunID: "run-001", Status: status,
		PromotionEligible: status == StatusPass,
		Observed: Observed{
			SemanticDigest: "sha256:semantic", SourceDigest: "sha256:source",
			Regions: regions, Slots: slots, Imports: imports, SourceMap: mappings,
			Delta: &Delta{Locality: []string{"billing.total", "billing.id"}},
		},
		Measurements: Measurements{RepeatCount: 2, CanonicalEqualCount: 2, SourceEqualCount: 2, SemanticEqualCount: 2, RegionEqualCount: 2},
		Evidence: EvidenceArtifact{Producer: "go", Bundle: EvidenceBundle{
			Schema: EvidenceSchema, Stage: StageGoBaseline, Fixture: "billing/main", Decision: string(status),
			Facts: []EvidenceFact{{ID: "billing/main/status", Kind: "status", Value: string(status)}, {ID: "billing/main/scope", Kind: "scope", Value: "billing"}},
		}},
	}
	if status == StatusFail {
		response.Failure = &Failure{Code: "marker-overlap", Detail: "protected marker changed"}
	}
	if reverse {
		response.Evidence.Bundle.Facts = reverseFacts(response.Evidence.Bundle.Facts)
	}
	return response
}

func reverseFacts(facts []EvidenceFact) []EvidenceFact {
	reversed := append([]EvidenceFact(nil), facts...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	return reversed
}
