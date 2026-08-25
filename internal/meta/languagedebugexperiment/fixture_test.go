package languagedebugexperiment

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"
)

func testInput(t *testing.T) Input {
	t.Helper()
	return Input{
		SubjectSHA: string(makeHex('a', 40)), ExecutableDigest: digest('d'),
		Contract:          fixedContract(),
		First:             languagedebug.Observe(executionFixture(t), "SOURCE_PARSED"),
		Second:            languagedebug.Observe(executionFixture(t), "ACTIVITY_INVOKED"),
		UnknownBreakpoint: languagedebug.Observe(executionFixture(t), "MISSING"),
	}
}

func executionFixture(t *testing.T) []byte {
	t.Helper()
	value := map[string]any{
		"schema": "gooo/source-execution-receipt/v1", "decision": "PASS", "resolution": "EXACT",
		"filename": "main.gooo", "source_digest": digest('a'), "semantic_digest": digest('b'),
		"entry": map[string]any{"activity": "PayOrder"}, "digest": digest('c'),
		"events": []map[string]any{{"sequence": 1, "kind": "SOURCE_PARSED", "subject": "a"},
			{"sequence": 2, "kind": "SEMANTIC_LOWERED", "subject": "b"},
			{"sequence": 3, "kind": "ACTIVITY_INVOKED", "subject": "PayOrder"},
			{"sequence": 4, "kind": "ENTITY_PRODUCED", "subject": "Receipt"}},
		"diagnostics": []any{}, "effects": map[string]any{"repository_writes": 0, "mutation_authority": false},
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func fixedContract() Contract {
	return Contract{"gooo/language-debug-experiment-contract/v1", 1, 2, 2, 2, 4, 2, 1, 2, 4, 2, 1, 3}
}

func digest(value byte) string { return "sha256:" + string(makeHex(value, 64)) }
func makeHex(value byte, size int) []byte {
	data := make([]byte, size)
	for index := range data {
		data[index] = value
	}
	return data
}
