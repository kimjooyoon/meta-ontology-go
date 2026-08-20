package analyzer

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"
)

func TestGeneratedBillingNormalizedDeltaIsStableAcrossRunsAndOrder(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	first := adaptGeneratedBillingSource(t, source, generatedBillingRegistry(t), policy)
	repeated := adaptGeneratedBillingSource(t, source, generatedBillingRegistry(t), policy)
	permuted := adaptGeneratedBillingSource(t, source, reversedGeneratedBillingRegistry(t), policy)

	for _, result := range []SemanticAdapterResult{first, repeated, permuted} {
		if len(result.NormalizedDelta.SignatureFacts) != 3 || len(result.NormalizedDelta.CandidateFacts) != 0 {
			t.Fatalf("normalized delta classes = %d signature, %d candidate",
				len(result.NormalizedDelta.SignatureFacts), len(result.NormalizedDelta.CandidateFacts))
		}
		if len(result.NormalizedDelta.DeferredImplementation) != 1 {
			t.Fatalf("deferred implementation observations = %d, want 1", len(result.NormalizedDelta.DeferredImplementation))
		}
		assertGeneratedDeltaBindings(t, result)
	}
	if first.NormalizedDelta.Digest != repeated.NormalizedDelta.Digest ||
		first.NormalizedDelta.Digest != permuted.NormalizedDelta.Digest {
		t.Fatal("normalized delta changed across repeat or registration-order permutation")
	}
	deferred := ReconcileSemantic(first, declaredBillingContract(t), first.SourceDigest,
		first.PolicyDigest, first.ToolchainDigest, "")
	if deferred.Accepted || deferred.WriteEffect != ReconcileNoWrite ||
		deferred.FailureCode != "source-or-observation-mismatch" {
		t.Fatalf("deferred implementation reconcile = %#v, want no-write", deferred)
	}
	payload, err := json.Marshal(first.NormalizedDelta)
	if err != nil {
		t.Fatalf("marshal normalized delta: %v", err)
	}
	for _, field := range []string{
		"schema_version", "signature_facts", "candidate_facts", "deferred_implementation",
		"deferred_details", "deferred_slots", "registry_digest", "digest",
	} {
		if !strings.Contains(string(payload), `"`+field+`"`) {
			t.Fatalf("machine-readable delta omitted %q: %s", field, payload)
		}
	}
}
func TestGeneratedBillingCandidateStaysNonAuthoritative(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	observed := adaptGeneratedBillingSource(t, source, ambiguousGeneratedBillingRegistry(t), policy)
	if len(observed.NormalizedDelta.CandidateFacts) != 1 {
		t.Fatalf("candidate facts = %d, want 1", len(observed.NormalizedDelta.CandidateFacts))
	}
	candidate := observed.NormalizedDelta.CandidateFacts[0]
	if len(candidate.Facts) != 2 || len(candidate.Evidence) != 2 {
		t.Fatalf("candidate handoff = %d facts, %d evidence; want two each", len(candidate.Facts), len(candidate.Evidence))
	}
	for _, fact := range candidate.Facts {
		if fact.Status != semantic.FactCandidate || observed.IR.Graph.HasFact(fact.Key()) ||
			!observed.IR.Graph.HasCandidate(fact.Key()) {
			t.Fatalf("candidate crossed authority boundary: %#v", fact)
		}
	}
	reconcile := ReconcileSemantic(observed, declaredBillingContract(t), observed.SourceDigest,
		observed.PolicyDigest, observed.ToolchainDigest, observed.ImplementationObservationDigest)
	if reconcile.Accepted || !reconcile.AuthoritySafe || reconcile.WriteEffect != ReconcileNoWrite {
		t.Fatalf("candidate reconcile = %#v, want no-write rejection", reconcile)
	}
}
