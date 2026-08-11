package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

const (
	maxGoFileLines     = 300
	maxGoFunctionLines = 75
)

func main() {
	root := flag.String("root", ".", "repository root")
	from := flag.String("from", os.Getenv("GOOO_SCOPE_FROM"), "base revision for scope checks")
	to := flag.String("to", os.Getenv("GOOO_SCOPE_TO"), "head revision for scope checks")
	branch := flag.String("branch", os.Getenv("GOOO_SCOPE_BRANCH"), "scope branch")
	head := flag.String("head", os.Getenv("GOOO_PR_HEAD"), "pull-request head branch")
	base := flag.String("base", os.Getenv("GOOO_PR_BASE"), "pull-request base branch")
	flag.Parse()
	if err := run(*root, *from, *to, *head, *base, *branch); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root, from, to, head, base, branch string) error {
	files, err := trackedGoFiles(root)
	if err != nil {
		return err
	}
	if err := verify.CheckGoCaps(root, files, maxGoFileLines, maxGoFunctionLines); err != nil {
		return err
	}
	if validRevision(from) && validRevision(to) && from != to {
		if !revisionAvailable(root, from) {
			return fmt.Errorf("scope base revision %q is unavailable", from)
		}
		if !revisionAvailable(root, to) {
			return fmt.Errorf("scope head revision %q is unavailable", to)
		}
		changed, err := changedPaths(root, from, to)
		if err != nil {
			return err
		}
		scopeBranch := branch
		if scopeBranch == "" {
			scopeBranch = head
		}
		if err := verify.CheckPathScopeForBranch(changed, scopeBranch); err != nil {
			return err
		}
		if scopeBranch == "agent/go-version" {
			diff, err := changedDiff(root, from, to, "go.mod")
			if err != nil {
				return err
			}
			if err := verify.CheckGoModToolchainDiff(diff); err != nil {
				return err
			}
		}
	}
	if base != "" || head != "" {
		if err := verify.CheckIntegrationPullRequest(head, base); err != nil {
			return err
		}
	}
	return nil
}

func validRevision(value string) bool {
	return value != "" && value != strings.Repeat("0", len(value))
}

func revisionAvailable(root, revision string) bool {
	if !validRevision(revision) {
		return false
	}
	_, err := runGit(root, "rev-parse", "--verify", revision+"^{commit}")
	return err == nil
}

func trackedGoFiles(root string) ([]string, error) {
	output, err := runGit(root, "ls-files", "-z", "--", "*.go")
	if err != nil {
		return nil, err
	}
	parts := strings.Split(output, "\x00")
	files := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			files = append(files, filepath.ToSlash(part))
		}
	}
	return files, nil
}

func changedPaths(root, from, to string) ([]string, error) {
	output, err := runGit(root, "diff", "--name-only", "--diff-filter=ACMRTUXB", from+"..."+to)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}

func changedDiff(root, from, to, path string) (string, error) {
	return runGit(root, "diff", "--unified=0", from+"..."+to, "--", path)
}

func runGit(root string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}
