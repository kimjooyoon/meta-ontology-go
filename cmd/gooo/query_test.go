package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestRunQueryReturnsStableSemanticProjection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runQuery([]string{"--json", "fixture.gooo", "--kind", "activity", "--predicate", "prov:used"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("query = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	var report jsonReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ok" || len(report.Nodes) != 1 || report.Nodes[0].Kind != "Activity" || len(report.Facts) != 1 {
		t.Fatalf("unexpected query projection: %#v", report)
	}
	if report.Facts[0].Predicate != "used" {
		t.Fatalf("predicate filter was not applied: %#v", report.Facts)
	}
}

func TestRunQueryUnknownIDHasStableFailureCode(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runQuery([]string{"--json", "fixture.gooo", "--id", "billing://entity/missing"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitFailure || stderr.Len() != 0 {
		t.Fatalf("unknown query ID = %d, stderr=%q", code, stderr.String())
	}
	var report jsonReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "error" || len(report.Diagnostics) != 1 || report.Diagnostics[0].Code != "query.invalid" {
		t.Fatalf("unexpected unknown-ID report: %#v", report)
	}
}

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
	var response queryengine.Response
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
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
		"--json", "billing.gooo", "--operation", "traverse",
		"--root", "billing://activity/pay-order", "--relation", "used",
		"--direction", "outgoing", "--layer", "deterministic",
		"--max-depth", "2", "--limit", "10",
	}
	first := runQueryBytes(t, traversalArgs, validSource)
	second := runQueryBytes(t, traversalArgs, validSource)
	if !bytes.Equal(first, second) {
		t.Fatalf("replayed traversal changed canonical output:\n%s\n%s", first, second)
	}
	var traversal queryengine.Response
	if err := json.Unmarshal(first, &traversal); err != nil {
		t.Fatal(err)
	}
	if len(traversal.Result.DeterministicPaths) != 1 || traversal.Status != queryengine.ResponseOK {
		t.Fatalf("traversal result = %#v", traversal)
	}

	derived := runQueryBytes(t, []string{
		"--json", "billing.gooo", "--derived",
		"--root", "billing://entity/order", "--rule", "usedBy",
		"--layer", "deterministic", "--max-depth", "1", "--limit", "10",
	}, validSource)
	var response queryengine.Response
	if err := json.Unmarshal(derived, &response); err != nil {
		t.Fatal(err)
	}
	if response.Status != queryengine.ResponseOK || len(response.Result.DerivedDeterministic) != 1 ||
		response.Result.DerivedDeterministic[0].Status != queryengine.DerivedFactStatus {
		t.Fatalf("derived result = %#v", response)
	}
}

func TestRunQueryEnvelopeRejectsUnknownAndIncompleteRequests(t *testing.T) {
	cases := []struct {
		name string
		args []string
		code string
	}{
		{
			name: "unknown predicate",
			args: []string{"--json", "billing.gooo", "--traverse", "--root", "billing://activity/pay-order", "--predicate", "prov:unknown", "--direction", "outgoing", "--limit", "10"},
			code: "unsupported_relation",
		},
		{
			name: "unknown endpoint",
			args: []string{"--json", "billing.gooo", "--traverse", "--root", "billing://activity/missing", "--direction", "outgoing", "--limit", "10"},
			code: "unknown_endpoint",
		},
		{
			name: "malformed rule",
			args: []string{"--json", "billing.gooo", "--derived", "--root", "billing://entity/order", "--rule", "not-a-rule", "--limit", "10"},
			code: "unsupported_rule",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runQuery(testCase.args, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
			if code != exitFailure || stderr.Len() != 0 {
				t.Fatalf("request = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
			}
			var response queryengine.Response
			if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Status != queryengine.ResponseError || response.Error == nil || response.Error.Code != testCase.code {
				t.Fatalf("response = %#v", response)
			}
		})
	}

	var stdout, stderr bytes.Buffer
	code := runQuery([]string{
		"--json", "billing.gooo", "--traverse", "--root", "billing://activity/pay-order",
		"--predicate", "used", "--direction", "outgoing", "--max-depth", "2", "--limit", "1",
	}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitFailure || stderr.Len() != 0 {
		t.Fatalf("incomplete query = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	var response queryengine.Response
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Status != queryengine.StatusDeferred || response.Error == nil || response.Error.Code != "incomplete_result" {
		t.Fatalf("incomplete response = %#v", response)
	}
}

func TestRunQueryEnvelopeUsesStableIDsAndPreservesCandidateIsolation(t *testing.T) {
	options := queryOptions{
		engine: true, operation: "derived", root: "billing://entity/order",
		rule: string(queryengine.RuleUsedBy), layer: string(queryengine.LayerAll),
		maxDepth: 1, maxDepthSet: true, limit: 10, limitSet: true,
	}
	ir := semantic.NewIR("billing", "billing")
	activity, err := semantic.NewActivity("billing://activity/pay", "billing", "PayOrder")
	if err != nil {
		t.Fatal(err)
	}
	order, err := semantic.NewEntity("billing://entity/order", "billing", "Order")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(order); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddCandidate(semantic.NewCandidateFact(activity.ID, semantic.Used, order.ID, "ambiguous")); err != nil {
		t.Fatal(err)
	}
	response, err := executeCLIQuery(ir, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DerivedDeterministic) != 0 || len(response.Result.DerivedCandidates) != 1 ||
		response.Result.DerivedCandidates[0].SourceLayer != queryengine.Candidate.String() {
		t.Fatalf("candidate entered authoritative result: %#v", response.Result)
	}

	graph, err := queryengine.FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
	request := queryRequest(options)
	if _, err := graph.Execute(request); err != nil {
		t.Fatal(err)
	}
	if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
		t.Fatal("CLI query execution mutated the detached authority projection")
	}
	if _, ok := graph.NodeByName("", "Order"); ok {
		t.Fatal("display name resolved without its namespace")
	}
	if !reflect.DeepEqual(response.Metadata.GraphHash, beforeHash) {
		t.Fatalf("response graph hash = %q, want %q", response.Metadata.GraphHash, beforeHash)
	}
}

func TestRunQueryEnvelopeCanonicalizesInputPermutation(t *testing.T) {
	args := []string{"--json", "billing.gooo", "--exact", "--root", "billing://activity/pay-order", "--relation", "used", "--target", "billing://entity/order", "--limit", "10"}
	first := runQueryBytes(t, args, validSource)
	permuted := `package billing
namespace billing
activity PayOrder(Order) -> Order
entity Order id "billing://entity/order"
`
	second := runQueryBytes(t, args, permuted)
	if !bytes.Equal(first, second) {
		t.Fatalf("input declaration permutation changed canonical response:\n%s\n%s", first, second)
	}
}

func runQueryBytes(t *testing.T, args []string, source string) []byte {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := runQuery(args, fixtureReader{source: source}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK || stderr.Len() != 0 {
		t.Fatalf("query = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	return stdout.Bytes()
}
