package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

func TestFixedRecordCLIFresh1763(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	base, err := os.ReadFile("../../examples/fixed-record-contract/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"complete", "one-expectation-mismatch", "unused-seed"} {
		t.Run(mode, func(t *testing.T) {
			source := string(base)
			if mode == "one-expectation-mismatch" {
				source = strings.Replace(source, `\"reason\":\"READY\"`, `\"reason\":\"WRONG\"`, 1)
			}
			if mode == "unused-seed" {
				source = strings.Replace(source, `attempts "1"`, "attempts \"1\"\nseed \"unnecessary\"", 1)
			}
			var stdout, stderr bytes.Buffer
			code := runBodyCodegen([]string{"--json", "--activity", "Decide", "fixed-record-1763.gooo"},
				fixtureReader{source: source}, &stdout, &stderr)
			if mode == "unused-seed" {
				if code != exitFailure || !strings.Contains(stdout.String(), "fixed value_case cannot combine") {
					t.Fatalf("fixed contract option diagnostic lost: code=%d %s %s", code, &stdout, &stderr)
				}
				return
			}
			if code != exitOK || stderr.Len() != 0 {
				t.Fatalf("real command did not generate fixed source: code=%d %s", code, &stderr)
			}
			var result bodycodegen.Result
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			r := result.Report.RecordAssembly
			passed, fields, status := 4, 12, "COMPLETE_FINITE"
			if mode == "one-expectation-mismatch" {
				passed, fields, status = 3, 11, "PARTIAL_FINITE"
			}
			if r == nil || r.Total != 4 || r.Passed != passed || r.FieldsTotal != 12 || r.FieldsPassed != fields ||
				r.Status != status || len(r.Choices) != 0 || len(r.Attempts) != 1 || len(r.Ranking) != 1 ||
				r.ModelRequested || r.ModelCalls != 0 || !result.Report.TypecheckPassed ||
				r.OriginalSourceSHA256 != projectionbundle.Digest([]byte(source)) ||
				result.Report.SourceDigest != r.SelectedSourceSHA256 {
				t.Fatalf("command changed source, finite scope or candidate count: %+v", r)
			}
		})
	}
}
