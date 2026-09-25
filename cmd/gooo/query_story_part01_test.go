package main

import "testing"

func TestParseQueryStoryArgumentsBindsReadOnlyLedger(t *testing.T) {
	options, filename, usage := parseQueryArguments([]string{
		"billing.gooo", "--story", "billing://activity/pay",
		"--provenance-ledger", "/tmp/billing-ledger.jsonl",
	})
	if usage != "" || filename != "billing.gooo" {
		t.Fatalf("story query parse failed: options=%#v filename=%q usage=%q", options, filename, usage)
	}
	if options.operation != "story" || options.storyID != "billing://activity/pay" || options.ledgerPath != "/tmp/billing-ledger.jsonl" {
		t.Fatalf("story query options = %#v", options)
	}
	if _, _, usage := parseQueryArguments([]string{"billing.gooo", "--story", "billing://activity/pay"}); usage == "" {
		t.Fatal("story query without a provenance ledger was accepted")
	}
	if _, _, usage := parseQueryArguments([]string{"billing.gooo", "--provenance-ledger", "/tmp/ledger"}); usage == "" {
		t.Fatal("provenance ledger without a story identity was accepted")
	}
}
