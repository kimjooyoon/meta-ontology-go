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
	head := flag.String("head", os.Getenv("GOOO_PR_HEAD"), "pull-request head branch")
	base := flag.String("base", os.Getenv("GOOO_PR_BASE"), "pull-request base branch")
	flag.Parse()
	if err := run(*root, *from, *to, *head, *base); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root, from, to, head, base string) error {
	files, err := trackedGoFiles(root)
	if err != nil {
		return err
	}
	if err := verify.CheckGoCaps(root, files, maxGoFileLines, maxGoFunctionLines); err != nil {
		return err
	}
	if from != "" && to != "" && from != to {
		changed, err := changedPaths(root, from, to)
		if err != nil {
			return err
		}
		if err := verify.CheckPathScope(changed, []string{".github", "scripts", "internal/verify"}); err != nil {
			return err
		}
	}
	if base != "" || head != "" {
		if err := verify.CheckIntegrationPullRequest(head, base); err != nil {
			return err
		}
	}
	return nil
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
	output, err := runGit(root, "diff", "--name-only", "--diff-filter=ACMRTUXB", from, to)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
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
