package main

import (
	"bytes"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"testing"
)

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
