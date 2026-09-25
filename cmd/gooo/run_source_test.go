package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

type runSourceReader struct{ source string }

func (reader runSourceReader) ReadFile(string) ([]byte, error) { return []byte(reader.source), nil }

const runSourceFixture = `package billing
namespace billing
entity In id "billing://entity/in"
entity Out id "billing://entity/out"
activity Execute(In) -> Out
`

func TestRunSourceJSONExecutesAndRejectsUnknownEntry(t *testing.T) {
	for _, test := range []struct {
		entry    string
		wantCode int
		decision string
	}{
		{"Execute", exitOK, "PASS"},
		{"Missing", exitFailure, "FAIL_CLOSED"},
	} {
		var stdout, stderr bytes.Buffer
		code := runSource([]string{"--json", "--entry", test.entry, "fixture.gooo"},
			runSourceReader{runSourceFixture}, &stdout, &stderr)
		var receipt sourceexecution.Receipt
		if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
			t.Fatal(err)
		}
		if code != test.wantCode || receipt.Decision != test.decision || stderr.Len() != 0 {
			t.Fatalf("code=%d receipt=%#v stderr=%q", code, receipt, stderr.String())
		}
	}
}
