package main

import (
	"bytes"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"testing"
)

func TestRunQueryEnvelopeExactBillingUsesDetachedQueryProjection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runQuery([]string{
		"--json", "billing.gooo", "--operation", "exact",
		"--root", "billing://activity/pay-order", "--relation", "prov:used",
		"--target", "billing://entity/order", "--layer", "deterministic",
		"--max-depth", "1", "--limit", "10",
	}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("exact query = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	response := decodeQueryResponse(t, stdout.Bytes())
	if response.Schema != queryengine.QueryEnvelopeSchema || response.Status != queryengine.ResponseOK {
		t.Fatalf("unexpected envelope identity: %#v", response)
	}
	if len(response.Result.DeterministicMatches) != 1 || response.Result.DeterministicMatches[0].Predicate != queryengine.Used {
		t.Fatalf("exact result = %#v", response.Result)
	}
	if response.Metadata.SemanticDigest == "" || response.Metadata.GraphHash == "" || response.Hash == "" {
		t.Fatalf("missing detached projection digests: %#v", response)
	}
	digest, err := response.CanonicalDigest()
	if err != nil || digest != response.Hash {
		t.Fatalf("canonical response digest = %q/%q, err=%v", digest, response.Hash, err)
	}
}
func TestRunQueryEnvelopeTraversalAndDerivedAreCanonicalAndBounded(t *testing.T) {
	traversalArgs := []string{
		"billing.gooo", "--operation", "traverse",
		"--root", "billing://activity/pay-order", "--relation", "used",
		"--direction", "outgoing", "--layer", "deterministic",
		"--max-depth", "2", "--limit", "10",
	}
	first := runQueryBytes(t, traversalArgs, validSource)
	second := runQueryBytes(t, traversalArgs, validSource)
	if !bytes.Equal(first, second) {
		t.Fatalf("replayed traversal changed canonical output:\n%s\n%s", first, second)
	}
	traversal := decodeQueryResponse(t, first)
	if len(traversal.Result.DeterministicPaths) != 1 || traversal.Status != queryengine.ResponseOK {
		t.Fatalf("traversal result = %#v", traversal)
	}

	derived := runQueryBytes(t, []string{
		"billing.gooo", "--derived",
		"--root", "billing://entity/order", "--rule", "usedBy",
		"--layer", "deterministic", "--max-depth", "1", "--limit", "10",
	}, validSource)
	response := decodeQueryResponse(t, derived)
	if response.Status != queryengine.ResponseOK || len(response.Result.DerivedDeterministic) != 1 ||
		response.Result.DerivedDeterministic[0].Status != queryengine.DerivedFactStatus {
		t.Fatalf("derived result = %#v", response)
	}
}
