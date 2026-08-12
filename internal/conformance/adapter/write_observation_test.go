package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOracleCorpusDeclaresExactStableCodes(t *testing.T) {
	data, err := os.ReadFile("testdata/no-write/oracle-corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name string `json:"name"`
		Code string `json:"oracle_code"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{
		OracleNW001: true, OracleNW002: true, OracleNW003: true, OracleNW004: true,
		OracleNW005: true, OracleNW006: true, OracleFAIL001: true, OracleFAIL002: true,
		OraclePASS001: true, OracleID001: true,
	}
	seen := make(map[string]bool, len(cases))
	for _, testCase := range cases {
		if testCase.Name == "" || !expected[testCase.Code] {
			t.Fatalf("invalid oracle fixture case: %+v", testCase)
		}
		seen[testCase.Code] = true
	}
	for code := range expected {
		if !seen[code] {
			t.Fatalf("oracle fixture corpus omitted %s", code)
		}
	}
}

func TestOracleNW003RejectsUntrustedOrMalformedObserverTrace(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	for _, test := range []struct {
		name string
		edit func(*NoWriteObservation)
	}{
		{name: "ORACLE-NW-003-untrusted", edit: func(observation *NoWriteObservation) { *observation = NoWriteObservation{} }},
		{name: "ORACLE-NW-003-malformed-digest", edit: func(observation *NoWriteObservation) { observation.Before.Source.ByteDigest = "sha256:bad" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			observer := newStableObserver(t, request)
			observation, err := observer.Finish()
			if err != nil {
				t.Fatal(err)
			}
			test.edit(&observation)
			evaluation := EvaluateObserved(request, response, &observation)
			assertOracleFailure(t, evaluation, OracleNW003)
		})
	}
}

func TestOracleNW002RejectsStaleObserverTrace(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(observation.Paths.OutputPath, []byte("changed after capture"), 0o600); err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, response, &observation)
	assertOracleFailure(t, evaluation, OracleNW002)
}

func TestOracleNW004RejectsSourceAndOutputByteChanges(t *testing.T) {
	for _, test := range []struct {
		name string
		path func(NoWriteObservation) string
	}{
		{name: "ORACLE-NW-004-source", path: func(observation NoWriteObservation) string { return observation.Paths.SourcePath }},
		{name: "ORACLE-NW-004-output", path: func(observation NoWriteObservation) string { return observation.Paths.OutputPath }},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := sampleRequest(StatusFail)
			request.Expected.FailureCode = "marker-overlap"
			observer := newStableObserver(t, request)
			paths := observer.paths
			if err := os.WriteFile(test.path(NoWriteObservation{Paths: paths}), []byte("different bytes"), 0o600); err != nil {
				t.Fatal(err)
			}
			observation, err := observer.Finish()
			if err != nil {
				t.Fatal(err)
			}
			evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
			assertOracleFailure(t, evaluation, OracleNW004)
		})
	}
}

func TestOracleNW005RejectsSameByteReplacementByLstat(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	original, err := os.ReadFile(observer.paths.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	replacement := filepath.Join(filepath.Dir(observer.paths.OutputPath), "replacement.go")
	if err := os.WriteFile(replacement, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, observer.paths.OutputPath); err != nil {
		t.Fatal(err)
	}
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleNW005)
}

func TestOracleNW006RejectsTempArtifactAddRemoveRenameAndMetadata(t *testing.T) {
	mutations := []struct {
		name string
		edit func(*testing.T, string)
	}{
		{name: "add", edit: func(t *testing.T, root string) { writeTemp(t, filepath.Join(root, "added.tmp"), "added") }},
		{name: "remove", edit: func(t *testing.T, root string) { removeTemp(t, filepath.Join(root, "stable.tmp")) }},
		{name: "rename", edit: func(t *testing.T, root string) {
			renameTemp(t, filepath.Join(root, "stable.tmp"), filepath.Join(root, "renamed.tmp"))
		}},
		{name: "metadata", edit: func(t *testing.T, root string) {
			if err := os.Chmod(filepath.Join(root, "stable.tmp"), 0o700); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, mutation := range mutations {
		t.Run("ORACLE-NW-006-"+mutation.name, func(t *testing.T) {
			request := sampleRequest(StatusFail)
			request.Expected.FailureCode = "marker-overlap"
			observer := newStableObserver(t, request)
			mutation.edit(t, observer.paths.TempRoot)
			observation, err := observer.Finish()
			if err != nil {
				t.Fatal(err)
			}
			evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
			assertOracleFailure(t, evaluation, OracleNW006)
		})
	}
}

func TestOracleNW006RejectsTempRootReplacement(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	oldRoot := observer.paths.TempRoot + ".old"
	if err := os.Rename(observer.paths.TempRoot, oldRoot); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(observer.paths.TempRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleNW006)
}

func TestOracleFAIL001RejectsMissingFailureProof(t *testing.T) {
	request := sampleRequest(StatusFail)
	response := sampleResponse(StatusFail, false)
	response.Failure = nil
	evaluation := Evaluate(request, response)
	assertOracleFailure(t, evaluation, OracleFAIL001)
}

func TestOracleFAIL002RejectsWrongFailureCode(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	response.Failure.Code = "different-failure"
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, response, &observation)
	assertOracleFailure(t, evaluation, OracleFAIL002)
}

func TestOraclePASS001RejectsProducerOnlyAndInvalidNoWriteClaims(t *testing.T) {
	request := sampleRequest(StatusPass)
	response := sampleResponse(StatusPass, false)
	claim := true
	response.ProducerClaims.NoWrite = &claim
	assertOracleFailure(t, Evaluate(request, response), OraclePASS001)
	invalid := NoWriteObservation{}
	assertOracleFailure(t, EvaluateObserved(request, response, &invalid), OraclePASS001)
}

func TestOracleID001RejectsRequestResponseAndObserverIdentityMismatch(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	response.RunID = "stale-run"
	assertOracleFailure(t, Evaluate(request, response), OracleID001)
	request.RunID = ""
	response.RunID = ""
	assertOracleFailure(t, Evaluate(request, response), OracleID001)
	request.RunID = "run-001"
	response.RunID = request.RunID
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	observation.Binding.Fixture = "other-fixture"
	assertOracleFailure(t, EvaluateObserved(request, response, &observation), OracleID001)
}

func newStableObserver(t *testing.T, request Request) *NoWriteObserver {
	return newObserverWithTempSetup(t, request, nil)
}

func newObserverWithTempSetup(t *testing.T, request Request, setup func(string)) *NoWriteObserver {
	t.Helper()
	root := t.TempDir()
	tempRoot := filepath.Join(root, "tmp")
	if err := os.Mkdir(tempRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(root, "source.gooo")
	outputPath := filepath.Join(root, "output.go")
	if err := os.WriteFile(sourcePath, []byte("entity billing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outputPath, []byte("package billing\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeTemp(t, filepath.Join(tempRoot, "stable.tmp"), "stable")
	if setup != nil {
		setup(tempRoot)
	}
	observer, err := NewNoWriteObserver(requestObservationBinding(request), ObserverPaths{
		SourcePath: sourcePath, OutputPath: outputPath, TempRoot: tempRoot,
	})
	if err != nil {
		t.Fatal(err)
	}
	attachVerifiedWorkflow(t, observer, request)
	return observer
}

func writeTemp(t *testing.T, path, value string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}

func removeTemp(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

func renameTemp(t *testing.T, oldPath, newPath string) {
	t.Helper()
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
}

func assertOracleFailure(t *testing.T, evaluation Evaluation, code string) {
	t.Helper()
	if evaluation.Matched || evaluation.OracleCode != code || evaluation.PromotionEligible {
		t.Fatalf("expected %s non-promotion failure, got %+v", code, evaluation)
	}
	if !strings.Contains(evaluation.Detail, code) {
		t.Fatalf("failure detail omitted %s: %q", code, evaluation.Detail)
	}
}
