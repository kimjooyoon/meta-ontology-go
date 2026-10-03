// Package bodyexecution independently compiles and executes a source-replayed
// pure typed-path projection. Model selection remains a preceding observation.
package bodyexecution

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ProcessObservation struct {
	Started          bool   `json:"started"`
	Completed        bool   `json:"completed"`
	Canceled         bool   `json:"canceled"`
	TimedOut         bool   `json:"timed_out"`
	DiagnosticsBytes int    `json:"diagnostics_bytes"`
	OutputTruncated  bool   `json:"output_truncated"`
	ExitCode         *int   `json:"exit_code"`
	WallNS           int64  `json:"wall_ns"`
	UserNS           int64  `json:"user_ns"`
	SystemNS         int64  `json:"system_ns"`
	PeakRSSBytes     *int64 `json:"peak_rss_bytes"`
	StdoutSHA256     string `json:"stdout_sha256"`
	StderrSHA256     string `json:"stderr_sha256"`
}

type limitedBuffer struct {
	buffer   bytes.Buffer
	limit    int
	exceeded bool
}

func (b *limitedBuffer) Len() int      { return b.buffer.Len() }
func (b *limitedBuffer) Bytes() []byte { return b.buffer.Bytes() }

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > b.limit {
		b.exceeded = true
		return 0, fmt.Errorf("child output exceeds observation bound")
	}
	return b.buffer.Write(p)
}

func process(ctx context.Context, dir, binary string, input []byte, args ...string) ([]byte, ProcessObservation, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	bindProcessGroup(cmd)
	cmd.Dir, cmd.WaitDelay = dir, time.Second
	cmd.Stdin = bytes.NewReader(input)
	cmd.Env = childEnvironment()
	stdout, stderr := &limitedBuffer{limit: 64 << 10}, &limitedBuffer{limit: 64 << 10}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	start := time.Now()
	err := cmd.Run()
	r := ProcessObservation{Started: cmd.Process != nil, Completed: err == nil, WallNS: time.Since(start).Nanoseconds(),
		Canceled: ctx.Err() == context.Canceled, TimedOut: ctx.Err() == context.DeadlineExceeded, DiagnosticsBytes: stderr.Len(),
		OutputTruncated: stdout.exceeded || stderr.exceeded,
		StdoutSHA256:    digest(stdout.Bytes()), StderrSHA256: digest(stderr.Bytes())}
	if cmd.ProcessState != nil {
		code := cmd.ProcessState.ExitCode()
		r.ExitCode = &code
		r.UserNS, r.SystemNS = cmd.ProcessState.UserTime().Nanoseconds(), cmd.ProcessState.SystemTime().Nanoseconds()
		r.PeakRSSBytes = peakRSS(cmd.ProcessState)
	}
	if err != nil || r.OutputTruncated {
		r.Completed = false
		return nil, r, fmt.Errorf("bounded child failed (exit/timeout recorded)")
	}
	if stderr.Len() != 0 {
		r.Completed = false
		return nil, r, fmt.Errorf("bounded child emitted unexpected diagnostics")
	}
	return stdout.Bytes(), r, nil
}

func childEnvironment() []string {
	var env []string
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		// Build and runtime children receive no model/provider configuration or
		// credentials. Only ordinary process/toolchain/cache locations are used.
		switch name {
		case "PATH", "HOME", "TMPDIR", "TMP", "TEMP", "SystemRoot", "SYSTEMROOT", "WINDIR", "USERPROFILE", "LOCALAPPDATA", "GOCACHE":
			env = append(env, entry)
		}
	}
	return append(env, "GOTOOLCHAIN=local", "GOENV=off", "GOWORK=off", "GO111MODULE=on", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0")
}
