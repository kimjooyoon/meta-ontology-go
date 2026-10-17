//go:build darwin || linux

package toolchaincli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestInvokeSeparatesExitAndOutputCompletion(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		exit int
		hold string
	}{{0, "none"}, {2, "none"}, {0, "both"}, {2, "both"}, {0, "stdout"}, {2, "stderr"}}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("exit_%d_hold_%s", tc.exit, tc.hold), func(t *testing.T) {
			root, release := t.TempDir(), ""
			if tc.hold != "none" {
				release = filepath.Join(root, "release-holder")
				t.Cleanup(func() { releaseInvocationHolder(t, release) })
			}
			arguments := []string{
				"-test.run=^TestInvocationOutputHelper$", "--", "emit",
				strconv.Itoa(tc.exit), release, tc.hold,
			}
			started := time.Now()
			session := Session{Executable: executable, Root: root}
			result, err := session.Invoke(arguments)
			if time.Since(started) > 8*time.Second {
				t.Fatal("output recovery exceeded the process and drain allowance")
			}
			validText := result.Stdout == "retained stdout" && result.Stderr == "retained stderr"
			if result.ExitCode != tc.exit || !validText {
				t.Fatal("observed parent exit or partial bytes disappeared", result, err)
			}
			if result.TreeBeforeDigest == "" || result.TreeBeforeDigest != result.TreeAfterDigest ||
				result.RepositoryWrites != 0 {
				t.Fatal("tree origins lost behind capture error", result, err)
			}
			if tc.hold == "none" {
				if err != nil || result.Failure != "" {
					t.Fatal("complete expected exit became incomplete", result, err)
				}
				return
			}
			capture, ok := errors.AsType[*OutputCaptureError](err)
			if !ok || !errors.Is(err, context.DeadlineExceeded) ||
				result.Failure != "INVOCATION_OUTPUT_INCOMPLETE" {
				t.Fatal("inherited pipe did not retain its independent capture cause", result, err)
			}
			for index, state := range []CaptureState{capture.Stdout, capture.Stderr} {
				open := tc.hold == "both" ||
					(index == 0 && tc.hold == "stdout") || (index == 1 && tc.hold == "stderr")
				if state.EOF == open || state.DeadlineCloseRequested != open ||
					state.LimitExceeded || state.RetainedBytes != 15 {
					t.Fatal("partial stream state was lost", index, state)
				}
			}
			if _, hasExit := errors.AsType[*exec.ExitError](err); hasExit != (tc.exit != 0) {
				t.Fatal("original nonzero exit cause was lost", err)
			}
		})
	}
}

func TestInvocationOutputHelper(t *testing.T) {
	var arguments []string
	for index, argument := range os.Args {
		if argument == "--" {
			arguments = os.Args[index+1:]
			break
		}
	}
	if len(arguments) == 0 {
		return
	}
	if arguments[0] == "hold" && len(arguments) == 2 {
		waitInvocationRelease(arguments[1])
		os.Exit(0)
	}
	if arguments[0] != "emit" || len(arguments) != 4 {
		t.Fatal("invalid invocation fixture arguments")
	}
	exit, err := strconv.Atoi(arguments[1])
	if err != nil {
		t.Fatal(err)
	}
	if arguments[2] != "" {
		helper := exec.Command(os.Args[0], "-test.run=^TestInvocationOutputHelper$", "--", "hold", arguments[2])
		if arguments[3] == "stdout" || arguments[3] == "both" {
			helper.Stdout = os.Stdout
		}
		if arguments[3] == "stderr" || arguments[3] == "both" {
			helper.Stderr = os.Stderr
		}
		if err := helper.Start(); err != nil {
			t.Fatal(err)
		}
	}
	_, _ = fmt.Fprint(os.Stdout, "retained stdout")
	_, _ = fmt.Fprint(os.Stderr, "retained stderr")
	os.Exit(exit)
}

func waitInvocationRelease(filename string) {
	deadline := time.NewTimer(12 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-deadline.C:
			return
		case <-tick.C:
			if _, err := os.Stat(filename); err == nil {
				_ = os.WriteFile(filename+".ack", []byte("release observed"), 0600)
				return
			}
		}
	}
}

func releaseInvocationHolder(t *testing.T, filename string) {
	t.Helper()
	if err := os.WriteFile(filename, []byte("release"), 0600); err != nil {
		t.Error(err)
		return
	}
	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-deadline.C:
			t.Error("fixture holder did not acknowledge release")
			return
		case <-tick.C:
			if _, err := os.Stat(filename + ".ack"); err == nil {
				return
			}
		}
	}
}
