package main

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestDurationUsesPositiveIntegerMilliseconds(t *testing.T) {
	value, err := durationMS("2026-08-30T00:00:00.000000001Z", "2026-08-30T00:00:00.000000002Z")
	if err != nil || value != 1 {
		t.Fatalf("duration = %d, err = %v", value, err)
	}
}

func TestObserveStepsIgnoresActionCleanupSteps(t *testing.T) {
	steps := []APIStep{
		{Name: "Run verification", Status: "completed", Conclusion: "success", StartedAt: "2026-08-30T00:00:00Z", CompletedAt: "2026-08-30T00:00:01Z"},
		{Name: "Post Run actions/setup-go@v6", Status: "completed", Conclusion: "success", StartedAt: "2026-08-30T00:00:01Z", CompletedAt: "2026-08-30T00:00:01Z"},
	}
	observed, total, err := observeSteps(steps)
	if err != nil || len(observed) != 1 || observed[0].Name != "Run verification" || total != 1000 {
		t.Fatalf("cleanup step was not excluded: steps=%+v total=%d err=%v", observed, total, err)
	}
}

func TestEqualNonOperationStepIsBoundedNotMissing(t *testing.T) {
	steps, total, err := observeSteps([]APIStep{{Name: "Complete job", Status: "completed", Conclusion: "success",
		StartedAt: "2026-08-30T00:00:00Z", CompletedAt: "2026-08-30T00:00:00Z"}})
	if err != nil || total != 0 || len(steps) != 1 || steps[0].Unknown != nil || !steps[0].BelowSourceResolution {
		t.Fatalf("equal timestamp was not bounded: steps=%+v total=%d err=%v", steps, total, err)
	}
}

func TestOperationBindsEvidenceAndGuardSteps(t *testing.T) {
	source := []byte("jobs:\n  check:\n    name: check\n    steps:\n      - name: Evidence\n        run: gofmt -l .\n      - name: Guard\n        run: test -s receipt.json\n")
	spec := OperationSpec{ID: "check", JobName: "check", StepName: "Guard", EvidenceStepName: "Evidence",
		GuardStepName: "Guard", Kind: "VERIFICATION", Command: []string{"gofmt", "-l", "."}, ProofObligationID: "ci-effort/check"}
	j := APIJob{ID: 1, Name: "check", Conclusion: "success", Steps: []APIStep{
		{Name: "Evidence", Status: "completed", Conclusion: "success", StartedAt: "2026-08-30T00:00:00Z", CompletedAt: "2026-08-30T00:00:01Z"},
		{Name: "Guard", Status: "completed", Conclusion: "success", StartedAt: "2026-08-30T00:00:01Z", CompletedAt: "2026-08-30T00:00:01Z"},
	}}
	operation := observeOperation(spec, []APIJob{j}, ".github/workflows/ci.yml", source, nil)
	if operation.State != "EXECUTED" || operation.EvidenceStepName != "Evidence" || !operation.GuardBound || operation.GuardStepName != "Guard" {
		t.Fatalf("evidence/guard binding was not preserved: %+v", operation)
	}
}

func TestEventSpecificOperationUsesMatchingPolicyStep(t *testing.T) {
	source := []byte("jobs:\n  policy:\n    name: policy\n    steps:\n      - name: Pull request policy\n        run: bash ./scripts/verify/validate-source-observations.sh $METRICS_DIR\n      - name: Push policy\n        run: bash ./scripts/verify/validate-source-observations.sh $METRICS_DIR\n")
	spec := OperationSpec{ID: "policy", JobName: "policy", StepName: "Push policy", EventStepNames: map[string]string{
		"pull_request": "Pull request policy", "push": "Push policy",
	}, Kind: "VERIFICATION", Command: []string{"bash", "./scripts/verify/validate-source-observations.sh", "$METRICS_DIR"}, ProofObligationID: "ci-effort/policy"}
	job := APIJob{ID: 1, Name: "policy", Conclusion: "success", Steps: []APIStep{{Name: "Pull request policy", Status: "completed", Conclusion: "success",
		StartedAt: "2026-08-30T00:00:00Z", CompletedAt: "2026-08-30T00:00:01Z"}}}
	operation := observeOperation(spec, []APIJob{job}, ".github/workflows/ci.yml", source, nil, "pull_request")
	if operation.State != "EXECUTED" || operation.EvidenceStepName != "Pull request policy" {
		t.Fatalf("event-specific policy step was not selected: %+v", operation)
	}
}

func TestUnknownEventDoesNotUseDeclaredFallbackStep(t *testing.T) {
	spec := OperationSpec{ID: "policy", JobName: "policy", StepName: "Push policy", EventStepNames: map[string]string{
		"pull_request": "Pull request policy", "push": "Push policy",
	}, Kind: "VERIFICATION", Command: []string{"bash", "check"}, ProofObligationID: "ci-effort/policy"}
	operation := observeOperation(spec, nil, ".github/workflows/ci.yml", []byte("jobs:\n"), nil, "workflow_dispatch")
	if operation.State != "UNKNOWN" || operation.EvidenceStepName != "" || operation.Unknown == nil || operation.Unknown.Reason != "EVENT_OPERATION_STEP_MISSING" {
		t.Fatalf("unknown event used a fallback step: %+v", operation)
	}
}

func TestMissingStepTimestampRemainsDirectUnknown(t *testing.T) {
	steps, _, err := observeSteps([]APIStep{{Name: "Verify", Status: "completed", Conclusion: "success"}})
	if err != nil || len(steps) != 1 || steps[0].Unknown == nil || steps[0].Unknown.Reason != "STEP_TIMESTAMP_MISSING" || steps[0].BelowSourceResolution {
		t.Fatalf("missing step timestamp was not direct unknown: steps=%+v err=%v", steps, err)
	}
}

func TestJobTimestampContradictionsAreRecorded(t *testing.T) {
	cases := []struct {
		name, started, completed, reason string
	}{
		{"malformed", "not-a-timestamp", "2026-08-30T00:00:01Z", "OPERATION_TIMESTAMP_MALFORMED"},
		{"negative", "2026-08-30T00:00:02Z", "2026-08-30T00:00:01Z", "OPERATION_DURATION_NEGATIVE"},
	}
	for _, test := range cases {
		observed, window, err := observeJobs([]APIJob{{ID: 1, Name: test.name, StartedAt: test.started, CompletedAt: test.completed}})
		if err != nil || len(observed) != 1 || observed[0].RejectionReason != test.reason || window.RuntimeRejectionCount != 1 || len(window.RuntimeRejectionReasons) != 1 || window.RuntimeRejectionReasons[0] != test.reason {
			t.Fatalf("%s timestamp contradiction was not recorded: jobs=%+v window=%+v err=%v", test.name, observed, window, err)
		}
	}
}

func TestStepTimestampContradictionsRefuteRuntime(t *testing.T) {
	cases := []struct {
		name, started, completed, reason string
	}{
		{"malformed", "not-a-timestamp", "2026-08-30T00:00:01Z", "OPERATION_TIMESTAMP_MALFORMED"},
		{"negative", "2026-08-30T00:00:02Z", "2026-08-30T00:00:01Z", "OPERATION_DURATION_NEGATIVE"},
	}
	for _, test := range cases {
		observed, window, err := observeJobs([]APIJob{{ID: 1, Name: test.name,
			StartedAt: "2026-08-30T00:00:00Z", CompletedAt: "2026-08-30T00:00:01Z",
			Steps: []APIStep{{Name: "Verify", Status: "completed", Conclusion: "success",
				StartedAt: test.started, CompletedAt: test.completed}}}})
		if err != nil || len(observed) != 1 || window.RuntimeRejectionCount != 1 || len(window.RuntimeRejectionReasons) != 1 || window.RuntimeRejectionReasons[0] != test.reason || observed[0].Steps[0].RejectionReason != test.reason {
			t.Fatalf("%s step timestamp contradiction was not recorded: jobs=%+v window=%+v err=%v", test.name, observed, window, err)
		}
		decision, resolution, reason := classifyReport(Report{SourceRunConclusion: "success", Window: window})
		if decision != "REFUTED" || resolution != "EXACT" || reason != "KNOWN_VERIFICATION_CONTRADICTION" {
			t.Fatalf("%s step timestamp contradiction was not refuted: %s/%s/%s", test.name, decision, resolution, reason)
		}
	}
}

func TestBoundZeroDurationRequiredStepIsTypedUnknown(t *testing.T) {
	source := []byte("jobs:\n  check:\n    name: check\n    steps:\n      - name: Verify\n        run: go test ./...\n")
	spec := OperationSpec{ID: "check", JobName: "check", StepName: "Verify", Kind: "VERIFICATION", Command: []string{"go", "test", "./..."}, ProofObligationID: "ci-effort/check"}
	jobs := []APIJob{{ID: 1, Name: "check", Conclusion: "success", Steps: []APIStep{{Name: "Verify", Status: "completed", Conclusion: "success", StartedAt: "2026-08-30T00:00:00Z", CompletedAt: "2026-08-30T00:00:00Z"}}}}
	operation := observeOperation(spec, jobs, ".github/workflows/ci.yml", source, nil)
	if operation.State != "UNKNOWN" || !validUnknown(operation.Unknown) || operation.Unknown.Reason != "OPERATION_DURATION_BELOW_SOURCE_RESOLUTION" || operation.Unknown.NextOperation != "OBSERVE_WITH_HIGHER_RESOLUTION_OR_REPEAT" {
		t.Fatalf("zero-duration bound step was not typed unknown: %+v", operation)
	}
}

func TestBoundMalformedTimestampIsRejected(t *testing.T) {
	source := []byte("jobs:\n  check:\n    name: check\n    steps:\n      - name: Verify\n        run: go test ./...\n")
	spec := OperationSpec{ID: "check", JobName: "check", StepName: "Verify", Kind: "VERIFICATION", Command: []string{"go", "test", "./..."}, ProofObligationID: "ci-effort/check"}
	jobs := []APIJob{{ID: 1, Name: "check", Conclusion: "success", Steps: []APIStep{{Name: "Verify", Status: "completed", Conclusion: "success", StartedAt: "not-a-timestamp", CompletedAt: "2026-08-30T00:00:01Z"}}}}
	operation := observeOperation(spec, jobs, ".github/workflows/ci.yml", source, nil)
	if operation.State != "REJECTED" || operation.RejectionReason != "OPERATION_TIMESTAMP_MALFORMED" {
		t.Fatalf("malformed bound timestamp was not rejected: %+v", operation)
	}
}

func TestBoundNegativeDurationIsRejected(t *testing.T) {
	source := []byte("jobs:\n  check:\n    name: check\n    steps:\n      - name: Verify\n        run: go test ./...\n")
	spec := OperationSpec{ID: "check", JobName: "check", StepName: "Verify", Kind: "VERIFICATION", Command: []string{"go", "test", "./..."}, ProofObligationID: "ci-effort/check"}
	jobs := []APIJob{{ID: 1, Name: "check", Conclusion: "success", Steps: []APIStep{{Name: "Verify", Status: "completed", Conclusion: "success", StartedAt: "2026-08-30T00:00:02Z", CompletedAt: "2026-08-30T00:00:01Z"}}}}
	operation := observeOperation(spec, jobs, ".github/workflows/ci.yml", source, nil)
	if operation.State != "REJECTED" || operation.RejectionReason != "OPERATION_DURATION_NEGATIVE" {
		t.Fatalf("negative bound duration was not rejected: %+v", operation)
	}
}

func TestReuseRequiresEveryContextDigest(t *testing.T) {
	base := ReuseKey{HeadSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", InputDigest: "input", ToolchainDigest: "toolchain", CommandContextDigest: "command", EnvironmentAllowlistDigest: "environment", DependencyGraphDigest: "dependency", ExpectedResultDigest: "expected", OpenTofuReleaseDigest: "release"}
	mutations := []func(*ReuseKey){
		func(key *ReuseKey) { key.HeadSHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" },
		func(key *ReuseKey) { key.InputDigest = "changed" },
		func(key *ReuseKey) { key.ToolchainDigest = "changed" },
		func(key *ReuseKey) { key.CommandContextDigest = "changed" },
		func(key *ReuseKey) { key.EnvironmentAllowlistDigest = "changed" },
		func(key *ReuseKey) { key.DependencyGraphDigest = "changed" },
		func(key *ReuseKey) { key.ExpectedResultDigest = "changed" },
		func(key *ReuseKey) { key.OpenTofuReleaseDigest = "changed" },
	}
	for index, mutate := range mutations {
		candidate := base
		mutate(&candidate)
		if sameReuseKey(base, candidate) {
			t.Fatalf("mutation %d was accepted", index)
		}
	}
}

func TestDependencyAbsenceIsStableAndDigestBound(t *testing.T) {
	missing := t.TempDir() + "/go.sum"
	first, err := readDependencyInputs([]string{missing})
	if err != nil || len(first) != 1 || first[0].State != "ABSENT" || first[0].Digest != "ABSENT" {
		t.Fatalf("missing dependency evidence = %+v, err = %v", first, err)
	}
	second, err := readDependencyInputs([]string{missing})
	if err != nil || digestJSON(first) != digestJSON(second) {
		t.Fatalf("repeated absence was not stable: first=%+v second=%+v err=%v", first, second, err)
	}
	if err := os.WriteFile(missing, []byte("module example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	present, err := readDependencyInputs([]string{missing})
	if err != nil || present[0].State != "PRESENT" || digestJSON(first) == digestJSON(present) {
		t.Fatalf("present dependency did not change evidence: absent=%+v present=%+v err=%v", first, present, err)
	}
}

func TestUnreadableDependencyIsNotAbsence(t *testing.T) {
	directory := t.TempDir() + "/dependency"
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := readDependencyInputs([]string{directory}); err == nil {
		t.Fatal("unreadable dependency was treated as absent")
	}
}

func TestMissingPriorHasCompleteUnknownContext(t *testing.T) {
	unknown := priorMissingUnknown()
	if !validUnknown(unknown) || unknown.UnknownClass != "DIRECT_MISSING" || len(unknown.BlockedBy) != 0 {
		t.Fatalf("unknown context is incomplete: %+v", unknown)
	}
}

func TestCounterexamplesPreserveReuseBoundary(t *testing.T) {
	cases := fixedCounterexamples()
	if len(cases) != 5 || cases[0].Decision != "REFUTED" || cases[2].Unknown == nil {
		t.Fatalf("counterexamples = %+v", cases)
	}
}

func TestWorkflowCommandBindingRejectsSourceCommandDrift(t *testing.T) {
	source := []byte("jobs:\n  check:\n    name: check\n    steps:\n      - name: Verify\n        run: go test ./...\n")
	spec := OperationSpec{ID: "check", JobName: "check", StepName: "Verify", Command: []string{"go", "test", "./..."}}
	if _, err := bindWorkflowCommand(source, ".github/workflows/ci.yml", spec); err != nil {
		t.Fatalf("valid workflow command rejected: %v", err)
	}
	mutated := []byte(strings.ReplaceAll(string(source), "go test ./...", "go vet ./..."))
	if _, err := bindWorkflowCommand(mutated, ".github/workflows/ci.yml", spec); err == nil {
		t.Fatal("workflow command drift was accepted")
	}
}

func TestEntityBindingPreservesOpenTofuAcronym(t *testing.T) {
	program := []byte("entity OpenTofuObservationInput id \"gooo://ci-effort-observation/input/opentofu-observation\"\n")
	if !hasEntityBinding(program, "gooo://ci-effort-observation/input/opentofu-observation") {
		t.Fatal("OpenTofu entity binding was rejected")
	}
	if hasEntityBinding(program, "gooo://ci-effort-observation/input/other") {
		t.Fatal("unrelated entity binding was accepted")
	}
}

func TestMissingCommandContextIsTypedUnknown(t *testing.T) {
	spec := OperationSpec{ID: "check", JobName: "check", StepName: "Verify", Kind: "VERIFICATION", Command: []string{"go", "test", "./..."}, ProofObligationID: "ci-effort/check"}
	operation := observeOperation(spec, nil, ".github/workflows/ci.yml", nil, errors.New("workflow source unavailable"))
	if operation.State != "UNKNOWN" || !validUnknown(operation.Unknown) || operation.Unknown.BlockedBy == nil {
		t.Fatalf("unknown operation lost causal context: %+v", operation)
	}
}

func TestMissingOpenTofuEvidenceIsTypedUnknown(t *testing.T) {
	value, err := observeOpenTofu("", "", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil || value.ArtifactID != 0 || !validUnknown(value.Unknown) {
		t.Fatalf("missing OpenTofu evidence = %+v, err = %v", value, err)
	}
}

func TestExactPriorReceiptAloneDischargesReuseObligation(t *testing.T) {
	digest := digestString("evidence")
	key := ReuseKey{HeadSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", InputDigest: digest,
		ToolchainDigest: digest, CommandContextDigest: digest, EnvironmentAllowlistDigest: digest,
		DependencyGraphDigest: digest, ExpectedResultDigest: digest, OpenTofuReleaseDigest: digest}
	prior := PriorRecord{Schema: reportSchema, Decision: "PASS", Resolution: "EXACT", HeadSHA: key.HeadSHA,
		Key: key, EvidenceDigest: digest, ResultDigest: digest, RepositoryWrites: 0}
	data, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/prior.json"
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	reuse, err := buildReuse(path, key)
	if err != nil || reuse.Decision != "REUSED" || reuse.Reused != 1 || reuse.RequiresExecution {
		t.Fatalf("exact prior reuse = %+v, err = %v", reuse, err)
	}
	key.InputDigest = digestString("changed")
	reuse, err = buildReuse(path, key)
	if err != nil || reuse.Decision != "REFUTED" || reuse.Reason != "REUSE_INPUT_DIGEST_MISMATCH" || reuse.RequiresExecution != true {
		t.Fatalf("mismatched prior reuse = %+v, err = %v", reuse, err)
	}
}

func TestRepositoryMutationIsNotAValidObservation(t *testing.T) {
	decision, resolution, reason := classifyReport(Report{SourceRunConclusion: "success", RepositoryWrites: 1})
	if decision != "REFUTED" || resolution != "EXACT" || reason != "KNOWN_VERIFICATION_CONTRADICTION" {
		t.Fatalf("repository mutation classified as %s/%s/%s", decision, resolution, reason)
	}
}
