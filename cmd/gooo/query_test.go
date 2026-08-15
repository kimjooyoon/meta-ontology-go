package main

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestRunQueryEnvelopeRejectsUnknownAndIncompleteRequests(t *testing.T) {
	cases := []struct {
		name string
		args []string
		code string
	}{
		{
			name: "unknown predicate",
			args: []string{"billing.gooo", "--traverse", "--root", "billing://activity/pay-order", "--predicate", "prov:unknown", "--direction", "outgoing", "--limit", "10"},
			code: "unsupported_relation",
		},
		{
			name: "unknown endpoint",
			args: []string{"billing.gooo", "--traverse", "--root", "billing://activity/missing", "--direction", "outgoing", "--limit", "10"},
			code: "unknown_endpoint",
		},
		{
			name: "malformed rule",
			args: []string{"billing.gooo", "--derived", "--root", "billing://entity/order", "--rule", "not-a-rule", "--limit", "10"},
			code: "unsupported_rule",
		},
		{
			name: "unrooted kind selector",
			args: []string{"billing.gooo", "--kind", "activity"},
			code: "invalid_root",
		},
		{
			name: "duplicate operation",
			args: []string{"billing.gooo", "--exact", "--traverse", "--root", "billing://activity/pay-order"},
			code: "unsupported_operation",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runQuery(testCase.args, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
			if code != exitFailure || stderr.Len() != 0 {
				t.Fatalf("request = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
			}
			response := decodeQueryResponse(t, stdout.Bytes())
			if response.Status != queryengine.ResponseError || response.Error == nil || response.Error.Code != testCase.code {
				t.Fatalf("response = %#v", response)
			}
			if response.Hash == "" || response.Hash != queryResponseDigestValue(response) {
				t.Fatalf("rejected response was not sealed: %#v", response)
			}
		})
	}

	var stdout, stderr bytes.Buffer
	code := runQuery([]string{
		"billing.gooo", "--traverse", "--root", "billing://activity/pay-order",
		"--predicate", "used", "--direction", "outgoing", "--max-depth", "2", "--limit", "1",
	}, fixtureReader{source: billingSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitFailure || stderr.Len() != 0 {
		t.Fatalf("incomplete query = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	response := decodeQueryResponse(t, stdout.Bytes())
	if response.Status != queryengine.StatusDeferred || response.Error == nil || response.Error.Code != "incomplete_result" {
		t.Fatalf("incomplete response = %#v", response)
	}
}

func TestRunQueryEnvelopeUsesStableIDsAndPreservesCandidateIsolation(t *testing.T) {
	options := queryOptions{
		operation: "derived", root: "billing://entity/order",
		rule: string(queryengine.RuleUsedBy), layer: string(queryengine.LayerDeterministic),
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
	if len(response.Result.DerivedDeterministic) != 0 || len(response.Result.DerivedCandidates) != 0 {
		t.Fatalf("candidate entered default authoritative result: %#v", response.Result)
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
	args := []string{"billing.gooo", "--exact", "--root", "billing://activity/pay-order", "--relation", "used", "--target", "billing://entity/order", "--limit", "10"}
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

func TestRunQueryDoesNotWriteAuthorityFile(t *testing.T) {
	directory := t.TempDir()
	filename := directory + "/billing.gooo"
	if err := os.WriteFile(filename, []byte(validSource), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	beforeEntries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := runQuery([]string{filename, "--id", "billing://activity/pay-order"}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("read-only query = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	after, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	afterEntries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) || !reflect.DeepEqual(beforeEntries, afterEntries) {
		t.Fatalf("query changed the authority filesystem: before=%v after=%v", beforeEntries, afterEntries)
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

func decodeQueryResponse(t *testing.T, payload []byte) queryengine.Response {
	t.Helper()
	var response queryengine.Response
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("query output was not canonical JSON: %v (%q)", err, payload)
	}
	if response.Schema != queryengine.QueryEnvelopeSchema {
		t.Fatalf("query output schema = %q, want %q", response.Schema, queryengine.QueryEnvelopeSchema)
	}
	return response
}

func queryResponseDigestValue(response queryengine.Response) string {
	digest, err := response.CanonicalDigest()
	if err != nil {
		return ""
	}
	return digest
}

const billingSource = `package billing
namespace billing
entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"
activity PayOrder(Order, PaymentMethod) -> Payment
`
