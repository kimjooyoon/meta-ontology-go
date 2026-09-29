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
	if code != exitFailure || !strings.Contains(stdout.String(), `"schema":"gooo/body-codegen-report/v3"`) || !strings.Contains(stdout.String(), `"decision":"FAIL_CLOSED"`) || !strings.Contains(stdout.String(), `"repository_writes":0`) || !strings.Contains(stdout.String(), `"completeness_receipt":{"schema":"gooo/metaprogramming-completeness-receipt/v2"`) || !strings.Contains(stdout.String(), `"fail_closed_reason":"`) {
		t.Fatalf("failed body-codegen = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunBodyCodegenAcceptsReplayableSampleSeed(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	valid := `package sample
namespace sample
entity Integer id "sample://entity/integer"
activity Choose(Integer) -> Integer computes "if input > 0 { return input } else { return 0 }"
`
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--json", "--sample-seed", "cli-replay-seed", "--activity", "Choose", "fixture.gooo"}, fixtureReader{source: valid}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"schema":"gooo/body-codegen-report/v3"`) ||
		!strings.Contains(stdout.String(), `"schema":"gooo/metaprogramming-completeness-receipt/v2"`) ||
		!strings.Contains(stdout.String(), `"aggregate_completeness_score":null`) ||
		!strings.Contains(stdout.String(), `"method":"sha256_seeded_weighted_choice/v1"`) ||
		!strings.Contains(stdout.String(), `"seed_sha256":"sha256:`) {
		t.Fatalf("seeded body-codegen = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), "cli-replay-seed") {
		t.Fatalf("raw sample seed leaked into JSON output: %s", stdout.String())
	}
}
