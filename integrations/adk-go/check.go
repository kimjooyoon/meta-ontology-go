// Package adkcheck exposes Gooo's deterministic source check as an optional ADK tool.
package adkcheck

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

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
	binary, root, timeout, err := resolveConfig(cfg)
	if err != nil {
		return nil, err
	}
	return functiontool.New(functiontool.Config{
		Name:        "gooo_check_source",
		Description: "Check one Gooo source file and return Gooo's unchanged JSON diagnostics.",
	}, func(ctx agent.Context, args CheckArgs) (CheckResult, error) {
		return check(ctx, binary, root, timeout, args)
	})
}

func resolveConfig(cfg Config) (string, string, time.Duration, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	binary, err := filepath.Abs(cfg.GoooBinary)
	if err != nil || cfg.GoooBinary == "" {
		return "", "", 0, failure(StageConfiguration, CodeInvalidBinary,
			errors.New("Gooo binary must be an absolute or resolvable path"))
	}
	root, err := filepath.Abs(cfg.SourceRoot)
	if err != nil || cfg.SourceRoot == "" {
		return "", "", 0, failure(StageConfiguration, CodeInvalidSourceRoot,
			errors.New("source root must be an absolute or resolvable path"))
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", "", 0, failure(StageConfiguration, CodeUnavailableRoot, fmt.Errorf("resolve source root: %w", err))
	}
	return binary, root, cfg.Timeout, nil
}
