package embeduse

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func TestPartitionCaptureRejectsChangedAndMislabeledBytes(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	name := "contract.gooo"
	const raw = "package p\nnamespace p\nentity A id \"urn:a\"\n"
	if err := os.WriteFile(filepath.Join(root, name), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	report := Report{Directory: root, Files: []File{{Member: sourcebundle.Member{
		Filename: name, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(raw))), ByteLength: len(raw)}}}}
	if _, err := captureInputs(report, nil); err != nil {
		t.Fatal(err)
	}
	projection := []sourcebundle.Source{{Filename: name, Content: raw}}
	if _, err := captureInputs(report, projection); err == nil {
		t.Fatal("observed disk bytes silently became a projection")
	}
	if err := os.WriteFile(filepath.Join(root, name), []byte(raw+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := captureInputs(report, nil); err == nil {
		t.Fatal("changed original source did not invalidate the observation")
	}
	report.Files[0].Projected = true
	if _, err := captureInputs(report, projection); err != nil {
		t.Fatal("exact projection did not retain its own source identity", err)
	}
}

func TestPartitionContextRejectsUnplannedSourceChanges(t *testing.T) {
	file := File{Member: sourcebundle.Member{Filename: "host.go", SHA256: "original", ByteLength: 5}}
	before := Report{Directory: "/same", ImportPath: "same", Files: []File{file}, Consumers: []Consumer{{Name: "input"}}}
	if err := samePartitionContext(before, before); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Report){
		func(r *Report) { r.Build = map[string]string{"GOFLAGS": "-tags=other"} },
		func(r *Report) { r.Files[0].SHA256 = "changed" },
		func(r *Report) { r.Files = nil },
		func(r *Report) { r.Files = append(r.Files, File{Member: sourcebundle.Member{Filename: "new.gooo"}}) },
		func(r *Report) { r.Consumers[0].Name = "other" },
		func(r *Report) { r.Consumers[0].Kind = "external-test" },
	} {
		after := before
		after.Files = append([]File(nil), before.Files...)
		after.Consumers = append([]Consumer(nil), before.Consumers...)
		mutate(&after)
		if err := samePartitionContext(before, after); err == nil {
			t.Fatal("unplanned host or member change was accepted")
		}
	}
}
