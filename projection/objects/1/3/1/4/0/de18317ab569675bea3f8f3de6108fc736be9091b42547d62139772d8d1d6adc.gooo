package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

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
