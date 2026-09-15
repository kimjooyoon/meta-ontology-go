package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func inspectFixtureOutput(t *testing.T, source string) []byte {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := runInspect([]string{"fixture.gooo"}, fixtureReader{source: source}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("inspect code = %d, stderr=%q", code, stderr.String())
	}
	return stdout.Bytes()
}
func decodeGraphDump(t *testing.T, output []byte) graphDump {
	t.Helper()
	var dump graphDump
	if err := json.Unmarshal(output, &dump); err != nil {
		t.Fatalf("decode graph dump: %v; output=%s", err, output)
	}
	return dump
}
func directoryEntries(t *testing.T, directory string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.Name())
	}
	return result
}
func hasGraphRelation(relations []graphRelation, status, subject, predicate, object string) bool {
	for _, relation := range relations {
		if relation.Status == status && relation.Subject == subject && relation.Predicate == predicate && relation.Object == object {
			return true
		}
	}
	return false
}

const sourceOrderA = `package billing
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder(Order) -> Payment
`
const sourceOrderB = `package billing
namespace billing
entity Payment id "billing://entity/payment"
entity Order id "billing://entity/order"
activity PayOrder(Order) -> Payment
`
