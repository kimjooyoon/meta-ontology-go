package toolchaincli

import (
	"errors"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMissingExecutablePreservesInvocationErrorAndTreeOrigins(t *testing.T) {
	root := t.TempDir()
	session := Session{Root: root, Executable: filepath.Join(root, "missing-executable")}
	arguments := []string{"version"}
	result, err := session.Invoke(arguments)
	if !errors.Is(err, fs.ErrNotExist) || result.ExitCode != -1 || result.Failure != "" {
		t.Fatal("launch failure became a resource observation", result, err)
	}
	if !reflect.DeepEqual(result.Arguments, arguments) || result.TreeBeforeDigest == "" ||
		result.TreeBeforeDigest != result.TreeAfterDigest || result.RepositoryWrites != 0 {
		t.Fatal("available invocation or tree origins were discarded", result)
	}
}

func TestSnapshotFailurePreservesAvailableOrigins(t *testing.T) {
	invocationErr := errors.New("invocation incomplete")
	snapshotErr := errors.New("after snapshot unavailable")
	for _, withInvocationError := range []bool{false, true} {
		name := "snapshot_only"
		var primary error
		if withInvocationError {
			name, primary = "invocation_and_snapshot", invocationErr
		}
		t.Run(name, func(t *testing.T) {
			observation := Observation{Arguments: []string{"check"}, ExitCode: -1,
				Stdout: "partial stdout", Stderr: "partial stderr", Failure: "INVOCATION_DEADLINE"}
			before := treeSnapshot{Digest: "known-before"}
			after := treeSnapshot{Digest: "unaccepted-partial-after"}
			result, err := completeInvocation(observation, before, after, primary, snapshotErr)
			want := observation
			want.TreeBeforeDigest = before.Digest
			if !reflect.DeepEqual(result, want) || !errors.Is(err, snapshotErr) ||
				errors.Is(err, invocationErr) != withInvocationError {
				t.Fatal("partial facts or original errors were discarded", result, err)
			}
		})
	}
}

func TestInvocationErrorRetainsCompletedTreeObservation(t *testing.T) {
	invocationErr := errors.New("invocation incomplete")
	before := treeSnapshot{Digest: "before", Records: map[string]treeRecord{
		"changed": {Path: "changed", ContentDigest: "old"},
	}}
	after := treeSnapshot{Digest: "after", Records: map[string]treeRecord{
		"changed": {Path: "changed", ContentDigest: "new"},
	}}
	after.Records["added"] = treeRecord{Path: "added", ContentDigest: "new"}
	observation := Observation{ExitCode: -1, Stderr: "partial stderr"}
	result, err := completeInvocation(observation, before, after, invocationErr, nil)
	want := observation
	want.TreeBeforeDigest, want.TreeAfterDigest, want.RepositoryWrites = "before", "after", 2
	if !reflect.DeepEqual(result, want) || err != invocationErr {
		t.Fatal("completed tree observation lost behind an invocation error", result, err)
	}
}
