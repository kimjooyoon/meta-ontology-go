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
		return CheckResult{}, err
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(callCtx, binary, "check", "--semantic", "--json", rel)
	cmd.Dir = root
	var stdout, stderr limitedBuffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if callCtx.Err() != nil {
		return CheckResult{}, callCtx.Err()
	}
	if errors.Is(stdout.err, errReportTooLarge) || errors.Is(stderr.err, errReportTooLarge) {
		return CheckResult{}, errReportTooLarge
	}
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			return CheckResult{}, fmt.Errorf("run Gooo check: %w", err)
		}
		code = exit.ExitCode()
	}
	if !json.Valid(stdout.Bytes()) {
		return CheckResult{}, fmt.Errorf("Gooo check returned invalid JSON (exit %d)", code)
	}
	return CheckResult{ReportJSON: stdout.String(), ExitCode: code}, nil
}
