package coupling

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCorpusJSONDoesNotContainExpectedLabels(t *testing.T) {
	row := testCorpus()[0]
	data, err := EncodeInputJSON(row.Input)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), row.Name) || strings.Contains(string(data), string(row.Expected.Reason)) {

		t.Fatal("fixture expectation escaped into input JSON")
	}
}
