package verify

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestProjectedCompileAdmissionFresh1873(t *testing.T) {
	for _, status := range []int{0, 23} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			root, command := projectedAdmissionFixture1873(t, status)
			err := command.Run()
			if status == 0 && err != nil || status != 0 && err == nil {
				t.Fatalf("projected command status changed: %v", err)
			}
			if status != 0 {
				exit, ok := errors.AsType[*exec.ExitError](err)
				if !ok || exit.ExitCode() != status {
					t.Fatalf("original failure status lost: %v", err)
				}
			}
			assertProjectedAdmission1873(t, root, status)
		})
	}
}

func projectedAdmissionFixture1873(t *testing.T, status int) (string, *exec.Cmd) {
	t.Helper()
	raw, err := os.ReadFile("../../scripts/semantic-conformance.sh")
	if err != nil {
		t.Fatal(err)
	}
	_, rest, found := strings.Cut(string(raw), "run_projected_compile() {\n")
	body, _, ended := strings.Cut(rest, "\n}\n")
	if !found || !ended {
		t.Fatal("projected compile entry point not found")
	}
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	fake := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$ARGUMENT_LOG\"\n" +
		"printf 'call\\n' >> \"$CALL_LOG\"\nprintf 'fixture-output\\n'\n" +
		"printf 'fixture-diagnostic\\n' >&2\nexit \"$FIXTURE_EXIT\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(fake), 0700); err != nil {
		t.Fatal(err)
	}
	harness := "set -uo pipefail\nprojection_work=\"$1\"\nrepo_root=\"$1\"\n" +
		"run_projected_compile() {\n" + body + "\n}\nrun_projected_compile\n"
	path := filepath.Join(root, "harness.sh")
	if err := os.WriteFile(path, []byte(harness), 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("bash", path, root)
	command.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"ARGUMENT_LOG="+filepath.Join(root, "arguments"), "CALL_LOG="+filepath.Join(root, "calls"),
		fmt.Sprintf("FIXTURE_EXIT=%d", status))
	return root, command
}

func assertProjectedAdmission1873(t *testing.T, root string, status int) {
	t.Helper()
	read := func(name string) []byte {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	var observed struct {
		Command      []string `json:"command"`
		CWD          string   `json:"cwd"`
		ExitCode     int      `json:"exit_code"`
		Observed     bool     `json:"observed"`
		StdoutBytes  int      `json:"stdout_bytes"`
		StdoutDigest string   `json:"stdout_digest"`
		StderrBytes  int      `json:"stderr_bytes"`
		StderrDigest string   `json:"stderr_digest"`
	}
	if err := json.Unmarshal(read("projected-compile.json"), &observed); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(observed.Command, []string{"go", "test", "-p=1", "./..."}) ||
		string(read("arguments")) != "test\n-p=1\n./...\n" || string(read("calls")) != "call\n" ||
		observed.ExitCode != status || observed.Observed != (status == 0) || observed.CWD != root {
		t.Fatal("scope, single admission or actual command identity changed", observed)
	}
	stdout, stderr := read("projected-compile.stdout"), read("projected-compile.stderr")
	if observed.StdoutBytes != len(stdout) || observed.StderrBytes != len(stderr) ||
		observed.StdoutDigest != fmt.Sprintf("sha256:%x", sha256.Sum256(stdout)) ||
		observed.StderrDigest != fmt.Sprintf("sha256:%x", sha256.Sum256(stderr)) {
		t.Fatal("original command stream origins changed", observed)
	}
}
