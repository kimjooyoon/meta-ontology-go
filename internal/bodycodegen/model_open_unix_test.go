//go:build unix

package bodycodegen

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/orderjudge"
)

func TestRetainedMetadataDoesNotWaitBeforeFileValidation(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "model.json")
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	checkModelFIFORejectsAndJoins(t, path, path, "bounded regular structural metadata")
}

func TestRetainedOrderWeightsDoNotWaitBeforeFileValidation(t *testing.T) {
	root := t.TempDir()
	path, fifo := filepath.Join(root, "model.json"), filepath.Join(root, "weights.bin")
	model, err := orderjudge.New([orderjudge.ParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	metadata, _, err := model.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, metadata, 0600); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	checkModelFIFORejectsAndJoins(t, path, fifo, "whole-candidate weights require a regular")
}

func checkModelFIFORejectsAndJoins(t *testing.T, model, fifo, diagnostic string) {
	t.Helper()
	done := make(chan error, 1)
	go func() { _, err := NewTypedPathGenerator(model); done <- err }()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), diagnostic) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		writer, err := os.OpenFile(fifo, os.O_WRONLY|syscall.O_NONBLOCK, 0)
		if err != nil {
			t.Fatal("cannot release model reader", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		select {
		case err := <-done:
			if err == nil || !strings.Contains(err.Error(), diagnostic) {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("model reader did not join after writer release")
		}
		t.Fatal("model file open waited before descriptor validation")
	}
}
