package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNamedReaderSavedSourceBinding(t *testing.T) {
	record, saved := readNamedReaderObservation(t)
	consumer := saved["sources/examples/source-control-reader/inspect.gooo"]
	if fmt.Sprintf("%x", sha256.Sum256(consumer)) != record.Source {
		t.Fatal("historical Gooo consumer source lost")
	}
	for _, path := range []string{"examples/named-inputs/main.gooo", "examples/source-control-reader/named-inputs.cases.json"} {
		source, err := os.ReadFile(filepath.Join("../..", path))
		if err != nil || !bytes.Equal(source, saved["sources/"+path]) {
			t.Fatal("saved reader input changed", path, err)
		}
	}
	assertNamedReaderResults(t, saved["native-transport-outputs/native-receipt.json"])
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(saved["prototype-field-assignment-repair-first/stdout"], &receipt); err != nil || receipt.Result == nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PASS" || receipt.Result.Runtime.FinitePassed != 16 ||
		receipt.Result.Runtime.FiniteTotal != 16 || receipt.Result.Runtime.ModelCalls != 0 {
		t.Fatal("saved finite observation scope changed")
	}
}

func readNamedReaderObservation(t *testing.T) (inputDeclarationObservation, map[string][]byte) {
	t.Helper()
	const directory = "../../docs/observations/named-input-reader"
	raw, err := os.ReadFile(filepath.Join(directory, "observations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record inputDeclarationObservation
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile(filepath.Join(directory, record.Archive.Path))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(archive)) != record.Archive.SHA256 {
		t.Fatal("saved reader archive changed", err)
	}
	z, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	return record, readInputDeclarationArchive(t, tar.NewReader(z), record.Archive.Members)
}
