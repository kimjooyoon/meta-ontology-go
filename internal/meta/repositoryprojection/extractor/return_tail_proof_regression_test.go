package extractor

import (
	"errors"
	"strings"
	"testing"
)

func TestReturnTailProofIdentityRegressionCohort(t *testing.T) {
	cases := []struct {
		name string
		run  func(*testing.T)
	}{
		{name: "P1 non-pass-proof-statuses-do-not-advance", run: testReturnTailNonPassProofStatuses},
		{name: "P2 proof-identities-bind-contract-source-semantic-and-candidate", run: testReturnTailProofIdentityBindings},
		{name: "P3 canonical-proof-fields-are-length-delimited", run: testReturnTailCanonicalProofFields},
		{name: "P4 renderer-selects-free-function-over-same-named-method", run: testReturnTailFreeFunctionRenderer},
		{name: "P5 generated-source-bytes-are-separate-from-evidence-payload", run: testReturnTailGeneratedByteAccounting},
	}
	if len(cases) != 5 {
		t.Fatalf("proof identity cohort denominator=%d, want 5", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, tc.run)
	}
}

func testReturnTailNonPassProofStatuses(t *testing.T) {
	for _, tc := range []struct {
		status string
		class  string
	}{
		{status: "UNKNOWN", class: "DIRECT_MISSING"},
		{status: "REFUTED", class: "KNOWN_CONTRADICTION"},
	} {
		chain := newReturnTailProofChain(singleReturnTailProofContract(), []byte("source"), []byte("candidate"), "contract-source", "contract-semantic")
		err := chain.consume(0, returnTailPredicateResult{Status: tc.status, Payload: []byte("payload"), CandidateDigest: proofDigest([]byte("candidate"))})
		var failure Failure
		if !errors.As(err, &failure) || failure.UnknownClass != tc.class || len(chain.stages) != 0 {
			t.Fatalf("status=%s failure=%v stages=%d", tc.status, err, len(chain.stages))
		}
	}
}

func testReturnTailProofIdentityBindings(t *testing.T) {
	base := returnTailProofOutputID([]byte("source"), []byte("candidate"), "contract-source", "contract-semantic", []byte("payload"))
	variants := []struct {
		name string
		id   string
	}{
		{name: "source", id: returnTailProofOutputID([]byte("source-2"), []byte("candidate"), "contract-source", "contract-semantic", []byte("payload"))},
		{name: "candidate", id: returnTailProofOutputID([]byte("source"), []byte("candidate-2"), "contract-source", "contract-semantic", []byte("payload"))},
		{name: "contract-source", id: returnTailProofOutputID([]byte("source"), []byte("candidate"), "contract-source-2", "contract-semantic", []byte("payload"))},
		{name: "contract-semantic", id: returnTailProofOutputID([]byte("source"), []byte("candidate"), "contract-source", "contract-semantic-2", []byte("payload"))},
		{name: "payload", id: returnTailProofOutputID([]byte("source"), []byte("candidate"), "contract-source", "contract-semantic", []byte("payload-2"))},
	}
	for _, variant := range variants {
		if variant.id == base {
			t.Fatalf("identity variant %s collided with base=%s", variant.name, base)
		}
	}

	first := returnTailProofOutputID([]byte("source"), []byte("candidate"), "contract-source", "contract-semantic", []byte("payload"))
	second := returnTailProofOutputID([]byte("source"), []byte("candidate"), "contract-source", "contract-semantic", []byte("payload"))
	if first != second {
		t.Fatalf("same proof inputs produced different identities: %s vs %s", first, second)
	}
}

func testReturnTailCanonicalProofFields(t *testing.T) {
	left := proofCanonical("ab", "c")
	right := proofCanonical("a", "bc")
	if strings.EqualFold(string(left), string(right)) {
		t.Fatalf("length-delimited proof fields collided: %q", left)
	}
}

func testReturnTailFreeFunctionRenderer(t *testing.T) {
	source := []byte("package p\n\ntype T struct{}\n\nfunc (T) F() error { return nil }\n\nfunc F() error { return nil }\n")
	rendered, err := renderedFunctionHelper(source, "F")
	if err != nil || !strings.Contains(string(rendered), "func F() error") || strings.Contains(string(rendered), "func (T) F() error") {
		t.Fatalf("renderer selected wrong declaration: err=%v rendered=%q", err, rendered)
	}
}

func testReturnTailGeneratedByteAccounting(t *testing.T) {
	generated := map[string][]byte{"b.go": []byte("package p\n"), "a.go": []byte("package p\n\nfunc F() error { return nil }\n")}
	if generatedSourceBytes(generated) != len(generated["a.go"])+len(generated["b.go"]) {
		t.Fatalf("source byte total=%d", generatedSourceBytes(generated))
	}
	if len(generatedPackagePayload(generated)) <= generatedSourceBytes(generated) {
		t.Fatalf("evidence payload did not retain canonical path framing")
	}
}

func singleReturnTailProofContract() []ContractObligationEvidence {
	return []ContractObligationEvidence{{
		Name: "return-shape", Activity: "ProveReturnShape", InputEntity: "FunctionInput", OutputEntity: "ReturnShapeObligation",
		UsedInputFact: true, GeneratedOutputFact: true,
	}}
}

func returnTailProofOutputID(source, candidate []byte, contractSource, contractSemantic string, payload []byte) string {
	chain := newReturnTailProofChain(singleReturnTailProofContract(), source, candidate, contractSource, contractSemantic)
	if err := chain.consume(0, returnTailPredicateResult{Status: "PASS", Payload: payload, CandidateDigest: proofDigest(candidate)}); err != nil {
		return "error:" + err.Error()
	}
	return chain.stages[0].OutputEvidenceID
}
