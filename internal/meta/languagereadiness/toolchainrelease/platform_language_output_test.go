package toolchainrelease

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLanguageSmokeRetainsFailedCommandBytes(t *testing.T) {
	for _, replay := range []bool{false, true} {
		t.Run(map[bool]string{false: "construct", true: "replay"}[replay], func(t *testing.T) {
			input := BuildInput{OutputDir: t.TempDir(), Target: Target{ID: "windows-amd64"}}
			failure := errors.New("native deadline exceeded")
			raw := []byte("{\"runtime\":{\"stage\":\"EXECUTE_1\"}}\n\xff\x00")
			run := func(string, []string, string, ...string) ([]byte, error) { return raw, failure }
			got, err := runLanguageSmokeCommand("gooo", "saved", input, languageSmokeCase{name: "text-operations"}, replay, run)
			mode := map[bool]string{false: "construct", true: "replay"}[replay]
			retained, readErr := os.ReadFile(filepath.Join(input.OutputDir, "windows-amd64-text-operations-"+mode+".failed-output"))
			if !errors.Is(err, failure) || readErr != nil || !bytes.Equal(got, raw) || !bytes.Equal(retained, raw) {
				t.Fatalf("failed command bytes/cause lost: %q %q %v %v", got, retained, err, readErr)
			}
		})
	}
}

func TestLanguageSmokeKeepsExecutionAndPersistenceFailures(t *testing.T) {
	input := BuildInput{OutputDir: filepath.Join(t.TempDir(), "absent"), Target: Target{ID: "windows-amd64"}}
	failure := errors.New("original command failure")
	run := func(string, []string, string, ...string) ([]byte, error) { return []byte("original"), failure }
	raw, err := runLanguageSmokeCommand("gooo", "saved", input, languageSmokeCase{name: "example"}, false, run)
	var pathErr *os.PathError
	if !errors.Is(err, failure) || !errors.As(err, &pathErr) || string(raw) != "original" {
		t.Fatal("one failure replaced the other", string(raw), err)
	}
}

func TestLanguageSmokeRetainsEmptyFailedOutput(t *testing.T) {
	input := BuildInput{OutputDir: t.TempDir(), Target: Target{ID: "windows-amd64"}}
	failure := errors.New("failed before output")
	run := func(string, []string, string, ...string) ([]byte, error) { return nil, failure }
	_, err := runLanguageSmokeCommand("gooo", "saved", input, languageSmokeCase{name: "example"}, true, run)
	retained, readErr := os.ReadFile(filepath.Join(input.OutputDir, "windows-amd64-example-replay.failed-output"))
	if !errors.Is(err, failure) || readErr != nil || len(retained) != 0 {
		t.Fatal("empty failed command observation missing", err, readErr)
	}
}

func TestLanguageSmokeRetainsSuccessfulJSONPath(t *testing.T) {
	input := BuildInput{OutputDir: t.TempDir(), Target: Target{ID: "darwin-arm64"}}
	raw := []byte("{\"generated_now\":true}\n")
	run := func(string, []string, string, ...string) ([]byte, error) { return raw, nil }
	_, err := runLanguageSmokeCommand("gooo", "saved", input, languageSmokeCase{name: "example"}, false, run)
	retained, readErr := os.ReadFile(filepath.Join(input.OutputDir, "darwin-arm64-example-construct.json"))
	if err != nil || readErr != nil || !bytes.Equal(retained, raw) {
		t.Fatal(err, readErr, string(retained))
	}
}
