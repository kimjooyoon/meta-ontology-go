package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func TestRevisionConsumerSnapshotDoesNotFollowCallerReplacement(t *testing.T) {
	original := filepath.Join(t.TempDir(), "consumer")
	data := []byte("explicit executable snapshot fixture, never executed")
	if err := os.WriteFile(original, data, 0600); err != nil {
		t.Fatal(err)
	}
	snapshot, digest, err := pinRevisionConsumer(t.TempDir(), original, policycompilation.DigestBytes(data))
	if err != nil || digest != policycompilation.DigestBytes(data) {
		t.Fatal("could not pin consumer fixture")
	}
	if err := os.WriteFile(original, []byte("replaced"), 0600); err != nil {
		t.Fatal(err)
	}
	captured, err := os.ReadFile(snapshot)
	if err != nil || !bytes.Equal(captured, data) {
		t.Fatal("caller replacement changed the executable snapshot")
	}
}

func TestRevisionConsumerOutputCaptureIsBounded(t *testing.T) {
	capture := &revisionBoundedCapture{limit: 3}
	if n, err := capture.Write([]byte("abcd")); n != 3 || err == nil || capture.String() != "abc" {
		t.Fatal("over-limit consumer output was accepted or discarded without a bound")
	}
}

func TestRevisionConsumerCopyCaptureIsBounded(t *testing.T) {
	for _, limit := range []int{0, 3, 4, 5} {
		for _, readerOnly := range []bool{false, true} {
			capture := &revisionBoundedCapture{limit: limit}
			var source io.Reader = strings.NewReader("abcd")
			if readerOnly {
				source = io.LimitReader(source, 4)
			}
			count, err := io.Copy(capture, source)
			wantCount := int64(min(limit, 4))
			if count != wantCount || (err != nil) != (limit < 4) || capture.String() != "abcd"[:wantCount] {
				t.Fatalf("limit=%d readerOnly=%t: count=%d retained=%q error=%v",
					limit, readerOnly, count, capture.String(), err)
			}
		}
	}
}

func TestRevisionConsumerProcessCaptureIsBounded(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestRevisionConsumerCaptureProcessHelper$")
	command.Env = append(os.Environ(), "GOOO_CAPTURE_TEST_CHILD=1")
	command.WaitDelay = time.Second
	capture := &revisionBoundedCapture{limit: 3}
	command.Stdout = capture
	if err := command.Run(); err == nil || ctx.Err() != nil || capture.String() != "abc" {
		t.Fatalf("retained=%q error=%v context=%v", capture.String(), err, ctx.Err())
	}
}

func TestRevisionConsumerCaptureProcessHelper(t *testing.T) {
	if os.Getenv("GOOO_CAPTURE_TEST_CHILD") != "1" {
		return
	}
	_, _ = io.WriteString(os.Stdout, "abcd")
	os.Exit(0)
}
