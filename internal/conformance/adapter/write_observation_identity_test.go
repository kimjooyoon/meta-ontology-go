package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestObserverCapturesIndependentRawBytesAndLstat(t *testing.T) {
	request := sampleRequest(StatusPass)
	observer := newStableObserver(t, request)
	source, err := os.Lstat(observer.paths.SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	got := observer.before.Source
	if got.ByteDigest != digestBytes([]byte("entity billing")) {
		t.Fatalf("source digest was not computed from raw bytes: %q", got.ByteDigest)
	}
	if !got.Lstat.Exists || got.Lstat.Mode != source.Mode().String() || got.Lstat.Size != source.Size() || got.Lstat.ModTimeUnixNano != source.ModTime().UnixNano() {
		t.Fatalf("source lstat metadata was not captured: %+v", got.Lstat)
	}
	if got.Lstat.Device == "" && got.Lstat.Inode == "" {
		t.Log("platform does not expose device/inode through FileInfo.Sys")
	}
}

func TestOracleNW006RejectsTempSymlinkIdentityChange(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	link := "stable.link"
	observer := newObserverWithTempSetup(t, request, func(root string) {
		if err := os.Symlink("stable.tmp", filepath.Join(root, link)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	})
	if err := os.Remove(filepath.Join(observer.paths.TempRoot, link)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("stable.tmp", filepath.Join(observer.paths.TempRoot, link)); err != nil {
		t.Fatal(err)
	}
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleNW006)
}

func TestOracleNW003RejectsObserverRelabelAndSnapshotTamper(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	request.RunID = "run-002"
	response.RunID = request.RunID
	observation.Binding.RunID = request.RunID
	evaluation := EvaluateObserved(request, response, &observation)
	assertOracleFailure(t, evaluation, OracleNW003)

	request = sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response = sampleResponse(StatusFail, false)
	observer = newStableObserver(t, request)
	observation, err = observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	observation.After.Output.ByteDigest = digestBytes([]byte("forged"))
	evaluation = EvaluateObserved(request, response, &observation)
	assertOracleFailure(t, evaluation, OracleNW003)
}

func TestOracleNW003RejectsWireRoundTripWithoutPrivateStamp(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(observation)
	if err != nil {
		t.Fatal(err)
	}
	var decoded NoWriteObservation
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &decoded)
	assertOracleFailure(t, evaluation, OracleNW003)
}
