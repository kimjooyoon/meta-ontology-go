package adkcheck

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

const maxQueryBytes = 512

// DiscoverArgs names source-owned capability declarations and a bounded query.
// DomainContract is optional and supplies a separate declaration denominator.
type DiscoverArgs struct {
	Path           string `json:"path"`
	Query          string `json:"query"`
	DomainContract string `json:"domain_contract,omitempty"`
}

// DiscoverResult preserves the compiler's complete JSON report and exit code.
type DiscoverResult struct {
	ReportJSON string `json:"report_json"`
	ExitCode   int    `json:"exit_code"`
}

// NewDiscover exposes deterministic capability discovery without model calls.
// The host fixes the binary and root; the agent chooses only a bounded query
// and source-root-relative .gooo files. This tool never executes cases.
func NewDiscover(cfg Config) (tool.Tool, error) {
	binary, root, timeout, err := resolveConfig(cfg)
	if err != nil {
		return nil, err
	}
	return functiontool.New(functiontool.Config{
		Name:        "gooo_discover_capability",
		Description: "Query declared Gooo capabilities for one source file; return the unchanged JSON receipt. A match is not implementation proof, and this tool never executes cases.",
	}, func(ctx agent.Context, args DiscoverArgs) (DiscoverResult, error) {
		result, err := discover(ctx, binary, root, timeout, args)
		return DiscoverResult(result), err
	})
}

func discover(ctx context.Context, binary, root string, timeout time.Duration, args DiscoverArgs) (CheckResult, error) {
	query := strings.TrimSpace(args.Query)
	if query == "" || len(query) > maxQueryBytes || strings.HasPrefix(query, "--") {
		return CheckResult{}, errors.New("query must contain 1 to 512 bytes and cannot begin with a flag")
	}
	rel, err := sourcePath(root, CheckArgs{Path: args.Path})
	if err != nil {
		return CheckResult{}, err
	}
	command := []string{"discover", "--json", "--query", query}
	if args.DomainContract != "" {
		contract, err := sourcePath(root, CheckArgs{Path: args.DomainContract})
		if err != nil {
			return CheckResult{}, err
		}
		command = append(command, "--domain-contract", contract)
	}
	return runJSONCommand(ctx, binary, root, timeout, append(command, rel)...)
}
