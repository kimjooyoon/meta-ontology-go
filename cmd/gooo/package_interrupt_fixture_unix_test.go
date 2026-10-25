//go:build unix

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

type packageInterruptFixture struct {
	command        *exec.Cmd
	stdout, stderr bytes.Buffer
	ready          string
	done           chan error
	pid            int
	joined         bool
}

func startPackageInterruptFixture(t *testing.T) *packageInterruptFixture {
	t.Helper()
	tool := filepath.Join(t.TempDir(), "go")
	script := "#!/bin/sh\nset -eu\nprintf '%s\\n' \"$$\" > \"$0.ready\"\nexec sleep 20\n"
	if err := os.WriteFile(tool, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	f := &packageInterruptFixture{ready: tool + ".ready", done: make(chan error, 1)}
	f.command = exec.Command(os.Args[0], "-test.run=^TestPackageExecuteInterruptJoinsOwnedTool$")
	f.command.Env = append(os.Environ(), "GOOO_INTERRUPT_TEST_CHILD=1", "GOOO_INTERRUPT_TEST_TOOL="+tool,
		"GOOO_LAYA_URL=", "GOOO_LAYA_API_KEY=")
	f.command.Stdout, f.command.Stderr = &f.stdout, &f.stderr
	if err := f.command.Start(); err != nil {
		t.Fatal(err)
	}
	go func() { f.done <- f.command.Wait() }()
	t.Cleanup(func() { f.close(t) })
	return f
}

func (f *packageInterruptFixture) awaitTool(t *testing.T) {
	t.Helper()
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		raw, _ := os.ReadFile(f.ready)
		f.pid, _ = strconv.Atoi(strings.TrimSpace(string(raw)))
		if f.pid > 0 {
			return
		}
		select {
		case err := <-f.done:
			f.joined = true
			t.Fatalf("CLI exited before its tool: %v stderr=%s stdout=%s", err, f.stderr.String(), f.stdout.String())
		case <-deadline.C:
			t.Fatal("tool did not reach readiness")
		case <-ticker.C:
		}
	}
}

func (f *packageInterruptFixture) close(t *testing.T) {
	t.Helper()
	raw, _ := os.ReadFile(f.ready)
	pid, _ := strconv.Atoi(strings.TrimSpace(string(raw)))
	if pid > 0 {
		if group, err := syscall.Getpgid(pid); err == nil && group == pid {
			_ = syscall.Kill(-pid, syscall.SIGKILL)
		} else {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	}
	if f.joined {
		return
	}
	_ = f.command.Process.Kill()
	select {
	case <-f.done:
		f.joined = true
	case <-time.After(3 * time.Second):
		t.Error("owned CLI did not join during cleanup")
	}
}
