package verticalsliceclosureshadow

import "fmt"

func toolchainFixture(head string) []byte {
	surfaces := []any{
		surfaceFixture("language-syntax-roundtrip", "gooo/language-syntax-roundtrip/v1", head, 20),
		surfaceFixture("language-semantic-model", "gooo/language-semantic-model/v1", head, 22),
		surfaceFixture("toolchain-executable-use-cases", "gooo/toolchain-executable-use-cases/v1", head, 3),
	}
	for index := range 6 {
		surfaces = append(surfaces, surfaceFixture(
			fmt.Sprintf("other-%d", index), "gooo/other/v1", head, 1))
	}
	return fixtureJSON(map[string]any{"schema": "gooo/toolchain-conformance-report/v1",
		"decision": "PASS", "resolution": "EXACT", "report_digest": fixtureDigest("e"),
		"repository_writes": 0, "mutation_authorized": false,
		"source": map[string]any{"expected_head_sha": head},
		"summary": map[string]any{"surfaces_satisfied": 9, "surfaces_total": 9,
			"cases_satisfied": 161, "cases_total": 161, "executed_cases": 161,
			"case_readiness_bps": 10000, "indicators_satisfied": 151,
			"indicators_total": 151, "proofs_passed": 27, "proofs_total": 27,
			"tamper_rejections": 13, "tamper_total": 13},
		"surfaces": surfaces})
}

func surfaceFixture(id, schema, head string, cases int) map[string]any {
	return map[string]any{"id": id, "schema": schema, "status": "SATISFIED",
		"head_sha": head, "cases": cases}
}

func releaseFixture(head string) []byte {
	cases := []any{}
	for _, target := range []string{"linux-amd64", "darwin-amd64", "windows-amd64"} {
		cases = append(cases, map[string]any{"id": target + "-go127-toolchain",
			"target_id": target, "observed": "go1.27.0", "expected": "go1.27.0"})
	}
	for index := range 17 {
		cases = append(cases, map[string]any{"id": fmt.Sprintf("case-%02d", index),
			"observed": "EXACT", "expected": "EXACT"})
	}
	return fixtureJSON(map[string]any{
		"schema":   "gooo/toolchain-cross-platform-release-report/v1",
		"decision": "PASS", "resolution": "EXACT", "head_sha": head,
		"report_digest": fixtureDigest("f"), "repository_writes": 0,
		"summary": map[string]any{"cases_satisfied": 20, "cases_total": 20,
			"readiness_bps": 10000, "platform_receipts": 3,
			"operating_systems": 3, "toolchain_bindings": 3},
		"cases": cases})
}
