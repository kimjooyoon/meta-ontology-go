package main

import (
	"bytes"
	"testing"

	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
)

func TestRunQueryLegacyIDAndExplicitExactShareCanonicalEnvelope(t *testing.T) {
	legacy := runQueryBytes(t, []string{
		"fixture.gooo", "--id", "billing://activity/pay-order",
		"--predicate", "prov:used", "--target", "billing://entity/order",
		"--layer", "deterministic", "--limit", "10",
	}, validSource)
	explicit := runQueryBytes(t, []string{
		"--json", "fixture.gooo", "--exact",
		"--root", "billing://activity/pay-order", "--relation", "used",
		"--target", "billing://entity/order", "--layer", "deterministic",
		"--max-depth", "1", "--limit", "10",
	}, validSource)
	if !bytes.Equal(legacy, explicit) {
		t.Fatalf("legacy and explicit exact requests diverged:\n%s\n%s", legacy, explicit)
	}
	response := decodeQueryResponse(t, legacy)
	if response.Status != queryengine.ResponseOK || response.Request.Operation != queryengine.OperationExact ||
		len(response.Result.DeterministicMatches) != 1 || response.Result.DeterministicMatches[0].Predicate != queryengine.Used {
		t.Fatalf("legacy exact response = %#v", response)
	}
}

func TestRunQueryLegacyIDSelectorUsesCanonicalTraversal(t *testing.T) {
	legacy := runQueryBytes(t, []string{
		"fixture.gooo", "--id", "billing://activity/pay-order",
	}, validSource)
	explicit := runQueryBytes(t, []string{
		"fixture.gooo", "--traverse", "--root", "billing://activity/pay-order",
		"--direction", "both", "--layer", "deterministic", "--max-depth", "1", "--limit", "100",
	}, validSource)
	if !bytes.Equal(legacy, explicit) {
		t.Fatalf("legacy ID traversal diverged from canonical traversal:\n%s\n%s", legacy, explicit)
	}
	response := decodeQueryResponse(t, legacy)
	if response.Request.Operation != queryengine.OperationTraversal || len(response.Result.DeterministicPaths) != 2 {
		t.Fatalf("legacy ID traversal = %#v", response)
	}
}

func TestRunQueryKindAndPredicateSelectorsUseOneCanonicalEnvelope(t *testing.T) {
	legacy := runQueryBytes(t, []string{
		"fixture.gooo", "--id", "billing://activity/pay-order",
		"--kind", "activity", "--predicate", "prov:used",
	}, validSource)
	explicit := runQueryBytes(t, []string{
		"fixture.gooo", "--traverse", "--root", "billing://activity/pay-order",
		"--relation", "used", "--direction", "both", "--layer", "deterministic",
		"--max-depth", "1", "--limit", "100",
	}, validSource)
	if !bytes.Equal(legacy, explicit) {
		t.Fatalf("selector-compatible request changed canonical envelope:\n%s\n%s", legacy, explicit)
	}
	response := decodeQueryResponse(t, legacy)
	if response.Status != queryengine.ResponseOK || len(response.Result.DeterministicPaths) != 1 ||
		response.Result.DeterministicPaths[0].Facts[0].Predicate != queryengine.Used {
		t.Fatalf("selector response = %#v", response)
	}
	if bytes.Contains(legacy, []byte(`"command"`)) || bytes.Contains(legacy, []byte(`"diagnostics"`)) {
		t.Fatalf("legacy CLI result schema remains reachable: %s", legacy)
	}
}
