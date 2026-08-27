package main

import (
	"bytes"
	"strings"
	"testing"
)

type invokeReader map[string][]byte

func (reader invokeReader) ReadFile(path string) ([]byte, error) {
	return reader[path], nil
}

func TestInvokeCommandExposesMetaProgramDecision(t *testing.T) {
	reader := invokeReader{
		"main.gooo": []byte(`package ciplan
namespace meta.ciplan
entity ChangeSet id "gooo://meta/ci-plan/entity/change-set"
entity CheckPlan id "gooo://meta/ci-plan/entity/check-plan"
entity VerificationReceipt id "gooo://meta/ci-plan/entity/verification-receipt"
activity SelectGoCheck(ChangeSet) -> CheckPlan computes "ci.rule:go:v1"
activity SelectDocsCheck(ChangeSet) -> CheckPlan computes "ci.rule:docs:v1"
activity SelectYAMLCheck(ChangeSet) -> CheckPlan computes "ci.rule:yaml:v1"
activity PlanCI(ChangeSet) -> CheckPlan computes "ci.plan:v1"
activity VerifyCIPlan(CheckPlan) -> VerificationReceipt computes "ci.verify:v1"
`),
		"input.json": []byte(`{"schema":"gooo/ci-plan-input/v1","case_id":"pass-go","files":["internal/a.go"]}`),
	}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runInvoke([]string{"--json", "--entry", "PlanCI", "--input", "input.json", "main.gooo"}, reader, stdout, stderr)
	if code != exitOK || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"decision": "PASS"`) {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestInvokeCommandReturnsFailureForLowerResolution(t *testing.T) {
	reader := invokeReader{
		"main.gooo": []byte(`package ciplan
namespace meta.ciplan
entity ChangeSet id "gooo://meta/ci-plan/entity/change-set"
entity CheckPlan id "gooo://meta/ci-plan/entity/check-plan"
entity VerificationReceipt id "gooo://meta/ci-plan/entity/verification-receipt"
activity SelectGoCheck(ChangeSet) -> CheckPlan computes "ci.rule:go:v1"
activity SelectDocsCheck(ChangeSet) -> CheckPlan computes "ci.rule:docs:v1"
activity SelectYAMLCheck(ChangeSet) -> CheckPlan computes "ci.rule:yaml:v1"
activity PlanCI(ChangeSet) -> CheckPlan computes "ci.plan:v1"
activity VerifyCIPlan(CheckPlan) -> VerificationReceipt computes "ci.verify:v1"
`),
		"input.json": []byte(`{"schema":"gooo/ci-plan-input/v1","case_id":"unknown","files":["LICENSE"]}`),
	}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runInvoke([]string{"--json", "--entry", "PlanCI", "--input", "input.json", "main.gooo"}, reader, stdout, stderr)
	if code != exitFailure || stderr.Len() != 0 || !strings.Contains(stdout.String(), `"resolution": "LOWER_RESOLUTION"`) {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}
