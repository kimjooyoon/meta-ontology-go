// Package adkcheck exposes Gooo's deterministic source check as an optional ADK tool.
package adkcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

const maxReportBytes = 1 << 20

// Config fixes the executable and source tree before an agent may choose a file.
type Config struct {
	GoooBinary string
	SourceRoot string
	Timeout    time.Duration
}

// CheckArgs names one source file relative to the configured source tree.
type CheckArgs struct {
	Path string `json:"path"`
}

// CheckResult carries the original Gooo JSON report and its process exit code.
type CheckResult struct {
	ReportJSON string `json:"report_json"`
	ExitCode   int    `json:"exit_code"`
}

// New returns a read-only ADK tool backed by the stable Gooo check command.
func New(cfg Config) (tool.Tool, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	binary, err := filepath.Abs(cfg.GoooBinary)
	if err != nil || cfg.GoooBinary == "" {
		return nil, errors.New("Gooo binary must be an absolute or resolvable path")
	}
	root, err := filepath.Abs(cfg.SourceRoot)
	if err != nil || cfg.SourceRoot == "" {
		return nil, errors.New("source root must be an absolute or resolvable path")
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve source root: %w", err)
	}
	return functiontool.New(functiontool.Config{
		Name:        "gooo_check_source",
		Description: "Check one Gooo source file and return Gooo's unchanged JSON diagnostics.",
	}, func(ctx agent.Context, args CheckArgs) (CheckResult, error) {
		return check(ctx, binary, root, cfg.Timeout, args)
	})
}

func check(ctx context.Context, binary, root string, timeout time.Duration, args CheckArgs) (CheckResult, error) {
	if args.Path == "" || filepath.IsAbs(args.Path) {
		return CheckResult{}, errors.New("path must be relative to the configured source root")
	}
	source, err := filepath.EvalSymlinks(filepath.Join(root, args.Path))
	if err != nil {
		return CheckResult{}, fmt.Errorf("resolve source path: %w", err)
	}
	rel, err := filepath.Rel(root, source)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return CheckResult{}, errors.New("source path escapes the configured root")
	}
	info, err := os.Stat(source)
	if err != nil || !info.Mode().IsRegular() || filepath.Ext(source) != ".gooo" {
		return CheckResult{}, errors.New("path must name a regular .gooo file")
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

var errReportTooLarge = errors.New("Gooo check output exceeds 1 MiB")

type limitedBuffer struct {
	bytes.Buffer
	err error
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > maxReportBytes {
		b.err = errReportTooLarge
		return 0, b.err
	}
	return b.Buffer.Write(p)
}
