package main

import (
	"bytes"
	"encoding/json"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"os"
	"reflect"
	"testing"
)

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
