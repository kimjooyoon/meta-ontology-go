package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

type revisionConsumerProcess struct {
	ExpectedExecutableDigest string `json:"expected_executable_digest"`
	ObservedExecutableDigest string `json:"observed_executable_digest"`
	Started                  bool   `json:"started"`
	ExitCode                 int    `json:"exit_code"`
	Stdout                   string `json:"stdout"`
	Stderr                   string `json:"stderr"`
	StdoutDigest             string `json:"stdout_digest"`
	StderrDigest             string `json:"stderr_digest"`
	WallMilliseconds         int64  `json:"wall_ms"`
}

func pinRevisionConsumer(directory, path, expected string) (string, string, error) {
	if !policycompilation.ValidDigest(expected) {
		return "", "", errors.New("invalid consumer executable digest")
	}
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", "", err
	}
	if !info.Mode().IsRegular() || info.Size() > 128<<20 {
		return "", "", errors.New("consumer must be a regular executable no larger than 128 MiB")
	}
	data, err := io.ReadAll(io.LimitReader(file, (128<<20)+1))
	if err != nil || len(data) > 128<<20 {
		return "", "", errors.Join(err, errors.New("consumer executable could not be captured within its bound"))
	}
	observed := policycompilation.DigestBytes(data)
	if observed != expected {
		return "", observed, errors.New("consumer executable digest mismatch")
	}
	name := "consumer"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	snapshot := filepath.Join(directory, name)
	if err := os.WriteFile(snapshot, data, 0500); err != nil {
		return "", observed, err
	}
	return snapshot, observed, nil
}

type revisionBoundedCapture struct {
	bytes.Buffer
	limit int
}

func (capture *revisionBoundedCapture) Write(data []byte) (int, error) {
	remaining := capture.limit - capture.Len()
	if len(data) > remaining {
		n, _ := capture.Buffer.Write(data[:remaining])
		return n, errors.New("consumer output exceeds its capture limit")
	}
	return capture.Buffer.Write(data)
}

func runRevisionConsumer(ctx context.Context, directory, executable string, source, request, report []byte,
	pkg, namespace string) (revisionConsumerProcess, error) {
	capture := revisionConsumerProcess{ExitCode: -1}
	paths := []string{}
	for index, name := range []string{"policy.gooo", "request.json", "report.json"} {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, [][]byte{source, request, report}[index], 0400); err != nil {
			return capture, err
		}
		paths = append(paths, path)
	}
	command := exec.CommandContext(ctx, executable, "-policy", paths[0], "-revision-request", paths[1],
		"-observe-revision-receipt", paths[2], "-profile-package", pkg, "-profile-namespace", namespace)
	command.Dir, command.WaitDelay = directory, time.Second
	stdout, stderr := &revisionBoundedCapture{limit: 16 << 20}, &revisionBoundedCapture{limit: 1 << 20}
	command.Stdout, command.Stderr = stdout, stderr
	start := time.Now()
	err := command.Run()
	capture.WallMilliseconds = time.Since(start).Milliseconds()
	capture.Started = command.ProcessState != nil
	if command.ProcessState != nil {
		capture.ExitCode = command.ProcessState.ExitCode()
	}
	capture.Stdout, capture.Stderr = stdout.String(), stderr.String()
	capture.StdoutDigest = policycompilation.DigestBytes(stdout.Bytes())
	capture.StderrDigest = policycompilation.DigestBytes(stderr.Bytes())
	if err != nil {
		return capture, fmt.Errorf("capture independent consumer: %w", err)
	}
	return capture, nil
}
