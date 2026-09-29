package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunBodyCodegenWritesProjectionOrClosedFailure(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	valid := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Clamp(Integer) -> Integer computes "if input > 0 { return input } else { return 0 }"
`
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--activity", "Clamp", "fixture.gooo"}, fixtureReader{source: valid}, &stdout, &stderr)
	if code != exitOK || !strings.Contains(stdout.String(), "func Clamp(input int64) int64") || stderr.Len() != 0 {
		t.Fatalf("body-codegen = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = runBodyCodegen([]string{"--json", "--activity", "Clamp", "fixture.gooo"}, fixtureReader{source: strings.Replace(valid, "return input", "return helper(input)", 1)}, &stdout, &stderr)
	if code != exitFailure || !strings.Contains(stdout.String(), `"schema":"gooo/body-codegen-report/v2"`) || !strings.Contains(stdout.String(), `"decision":"FAIL_CLOSED"`) || !strings.Contains(stdout.String(), `"repository_writes":0`) {
		t.Fatalf("failed body-codegen = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}
