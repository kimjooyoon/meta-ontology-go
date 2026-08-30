package governancesnapshot

import (
	"encoding/json"
	"fmt"
)

func canonicalCases(input LoadedSnapshot, contract Contract) []CanonicalCase {
	current := evaluateAndFinish(input, contract)
	normal := evaluateAndFinish(fixtureSnapshot(contract, "normal"), contract)
	missing := evaluateAndFinish(fixtureSnapshot(contract, "missing"), contract)
	dependency := evaluateAndFinish(fixtureSnapshot(contract, "dependency"), contract)
	malformed := evaluateAndFinish(fixtureSnapshot(contract, "malformed"), contract)
	disabled := evaluateAndFinish(fixtureSnapshot(contract, "disabled"), contract)
	return []CanonicalCase{
		caseFromReport("normal-main-match", normal),
		caseFromReport("current-dev-drift", current),
		caseFromReport("missing-public-snapshot", missing),
		caseFromReport("dependency-blocked-context-comparison", dependency),
		caseFromReport("malformed-payload", malformed),
		caseFromReport("disabled-ruleset-authority", disabled),
	}
}

func evaluateAndFinish(input LoadedSnapshot, contract Contract) Report {
	report := evaluateCore(input, contract)
	finishReport(&report, contract)
	return report
}

func caseFromReport(id string, report Report) CanonicalCase {
	result := CanonicalCase{ID: id, Decision: report.Decision, Resolution: report.Resolution, Reason: report.Reason}
	for _, cell := range report.Cells {
		if cell.Decision == DecisionRefuted {
			result.Counterexample = cell.Counterexample
			result.Expected = cell.Expected
			result.Observed = cell.Observed
			return result
		}
	}
	for _, cell := range report.Cells {
		if cell.Unknown != nil {
			unknown := *cell.Unknown
			result.Unknown = &unknown
			result.Expected = cell.Expected
			result.Observed = cell.Observed
			return result
		}
	}
	return result
}

func fixtureSnapshot(contract Contract, mode string) LoadedSnapshot {
	result := LoadedSnapshot{Snapshot: Snapshot{Schema: SnapshotSchema, Repository: contract.Expected.Repository,
		HeadSHA: "fixture-head"}, Payloads: map[string][]byte{}}
	for _, endpoint := range contract.Source.Endpoints {
		request := RequestObservation{ID: endpoint.ID, Method: endpoint.Method, URL: endpoint.Path,
			APIVersion: endpoint.API, PayloadPath: endpoint.Payload, State: "PRESENT"}
		if mode == "missing" {
			request.State = "MISSING"
		}
		if mode == "dependency" && request.ID == "main-protection" {
			request.State = "MISSING"
		}
		result.Requests = append(result.Requests, request)
	}
	result.Payloads["repository"] = []byte(`{"default_branch":"dev"}`)
	result.Payloads["dev-branch"] = []byte(`{"name":"dev","protected":true,"commit":{"sha":"dev-fixture"}}`)
	result.Payloads["main-branch"] = []byte(`{"name":"main","protected":true,"commit":{"sha":"main-fixture"}}`)
	result.Payloads["dev-protection"] = protectionFixture()
	result.Payloads["main-protection"] = protectionFixture()
	result.Payloads["rulesets"] = rulesetsFixture(contract, false)
	if mode == "drift" {
		result.Payloads["dev-protection"] = []byte(`{"required_status_checks":null}`)
	}
	if mode == "malformed" {
		result.Payloads["main-branch"] = []byte("{")
	}
	if mode == "disabled" {
		result.Payloads["rulesets"] = rulesetsFixture(contract, true)
	}
	for index := range result.Requests {
		request := &result.Requests[index]
		if request.State == "PRESENT" {
			request.PayloadDigest, _ = normalizedDigest(result.Payloads[request.ID])
		}
	}
	return result
}

func protectionFixture() []byte {
	data := map[string]any{"required_status_checks": map[string]any{"enforcement_level": "everyone", "contexts": ExpectedContexts(), "checks": []any{}}}
	raw, _ := json.Marshal(data)
	return raw
}

func rulesetsFixture(contract Contract, authority bool) []byte {
	values := make([]map[string]any, 0, len(contract.Expected.Rulesets))
	for _, expected := range contract.Expected.Rulesets {
		values = append(values, map[string]any{"id": expected.ID, "name": expected.Name, "target": "branch", "enforcement": expected.Enforcement, "authority": authority})
	}
	raw, _ := json.Marshal(values)
	return raw
}

func validateRequests(input LoadedSnapshot, contract Contract) string {
	if len(input.Requests) != len(contract.Source.Endpoints) {
		return "SOURCE_REQUEST_COUNT_MISMATCH"
	}
	for index, expected := range contract.Source.Endpoints {
		request := input.Requests[index]
		if request.ID != expected.ID || request.Method != expected.Method || request.URL != expected.Path || request.APIVersion != expected.API || request.PayloadPath != expected.Payload {
			return fmt.Sprintf("SOURCE_REQUEST_MISMATCH:%s", expected.ID)
		}
		if request.State != "PRESENT" && request.State != "MISSING" && request.State != "UNAVAILABLE" {
			return "SOURCE_REQUEST_STATE_INVALID"
		}
		if request.State == "PRESENT" {
			raw := input.Payloads[request.ID]
			digest, err := normalizedDigest(raw)
			if err != nil {
				return "MALFORMED_PUBLIC_PAYLOAD"
			}
			if request.PayloadDigest != digest {
				return "PAYLOAD_DIGEST_MISMATCH"
			}
		}
	}
	return ""
}
