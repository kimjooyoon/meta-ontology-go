package adkcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

func check(ctx context.Context, binary, root string, timeout time.Duration, args CheckArgs) (CheckResult, error) {
	rel, err := sourcePath(root, args)
	if err != nil {
		return CheckResult{}, failure(StageInput, CodeInvalidPath, err)
	}
	return runJSONCommand(ctx, binary, root, timeout, "check", "--semantic", "--json", rel)
}

func runJSONCommand(ctx context.Context, binary, root string, timeout time.Duration, args ...string) (CheckResult, error) {
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(callCtx, binary, args...)
	cmd.Dir = root
	var stdout, stderr limitedBuffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if callCtx.Err() != nil {
		code := CodeDeadline
		if errors.Is(callCtx.Err(), context.Canceled) {
			code = CodeCancelled
		}
		return CheckResult{}, failure(StageExecution, code, callCtx.Err())
	}
	if errors.Is(stdout.err, errReportTooLarge) || errors.Is(stderr.err, errReportTooLarge) {
		return CheckResult{}, failure(StageOutput, CodeOutputLimit, errReportTooLarge)
	}
	code := 0
	if err != nil {
		exit, ok := errors.AsType[*exec.ExitError](err)
		if !ok {
			return CheckResult{}, failure(StageExecution, CodeStartFailed, fmt.Errorf("run Gooo: %w", err))
		}
		code = exit.ExitCode()
		if code < 0 {
			return CheckResult{}, failure(StageExecution, CodeTerminated, err)
		}
	}
	if !json.Valid(stdout.Bytes()) {
		return CheckResult{}, failure(StageOutput, CodeInvalidJSON,
			fmt.Errorf("Gooo returned invalid JSON (exit %d)", code))
	}
	return CheckResult{ReportJSON: stdout.String(), ExitCode: code}, nil
}
