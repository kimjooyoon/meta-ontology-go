package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func main() {
	root := flag.String("root", ".", "repository root")
	from := flag.String("from", os.Getenv("GOOO_SCOPE_FROM"), "base revision for scope checks")
	to := flag.String("to", os.Getenv("GOOO_SCOPE_TO"), "head revision for scope checks")
	branch := flag.String("branch", os.Getenv("GOOO_SCOPE_BRANCH"), "scope branch")
	head := flag.String("head", os.Getenv("GOOO_PR_HEAD"), "pull-request head branch")
	base := flag.String("base", os.Getenv("GOOO_PR_BASE"), "pull-request base branch")
	expectedHead := flag.String("expected-head", os.Getenv("GOOO_EXPECTED_HEAD"), "expected checked-out pull-request head revision")
	capsOnly := flag.Bool("caps-only", false, "run only DAMP/DRY caps")
	skipCaps := flag.Bool("skip-caps", false, "skip DAMP/DRY caps and run scope checks")
	flag.Parse()
	if err := validateCapMode(*capsOnly, *skipCaps); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := run(*root, *from, *to, *head, *base, *branch, *expectedHead, *capsOnly, *skipCaps); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func validateCapMode(capsOnly, skipCaps bool) error {
	if capsOnly && skipCaps {
		return fmt.Errorf("--caps-only and --skip-caps are mutually exclusive")
	}
	return nil
}

func run(root, from, to, head, base, branch, expectedHead string, capsOnly, skipCaps bool) error {
	if !skipCaps {
		policy := verify.DefaultLinePolicy()
		if err := verify.CheckSourcePolicy(root, nil, policy); err != nil {
			return err
		}
	}
	if capsOnly {
		return nil
	}
	if err := checkAgentPushBranch(branch); err != nil {
		return err
	}
	if err := validateScopeRevisions(from, to, expectedHead); err != nil {
		return err
	}
	if err := verifyPRCheckoutIdentity(root, from, to, expectedHead); err != nil {
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
		promotion := base == "main" && head == "dev"
		if !promotion {
			if err := verify.CheckPathScopeForBranch(changed, scopeBranch); err != nil {
				return err
			}
		}
		if scopeBranch == "agent/go-version" && !promotion {
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
		if err := verify.CheckPullRequestPolicy(head, base); err != nil {
			return err
		}
	}
	return nil
}

func validateScopeRevisions(from, to, expectedHead string) error {
	if from == "" && to == "" {
		if expectedHead != "" {
			return fmt.Errorf("pull-request scope revisions must not be empty")
		}
		return nil
	}
	if !validRevision(from) || !validRevision(to) {
		return fmt.Errorf("scope revisions must both be valid or both be empty")
	}
	if from == to {
		return fmt.Errorf("scope base and head revisions must differ")
	}
	if expectedHead != "" && to != expectedHead {
		return fmt.Errorf("scope head %q does not match expected PR head %q", to, expectedHead)
	}
	return nil
}

func checkAgentPushBranch(branch string) error {
	if !strings.HasPrefix(branch, "agent/") {
		return nil
	}
	return verify.CheckPathScopeForBranch(nil, branch)
}

func verifyPRCheckoutIdentity(root, base, to, expectedHead string) error {
	if expectedHead == "" {
		return nil
	}
	actualHead, err := runGit(root, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	actualHead = strings.TrimSpace(actualHead)
	if actualHead != expectedHead {
		return fmt.Errorf("checked out revision %q does not match expected PR head %q", actualHead, expectedHead)
	}
	if to != expectedHead {
		return fmt.Errorf("scope head %q does not match expected PR head %q", to, expectedHead)
	}
	if !revisionAvailable(root, base) {
		return fmt.Errorf("scope base revision %q is unavailable", base)
	}
	return nil
}

func validRevision(value string) bool {
	if len(value) != 40 || value == strings.Repeat("0", 40) {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}

func revisionAvailable(root, revision string) bool {
	if !validRevision(revision) {
		return false
	}
	_, err := runGit(root, "rev-parse", "--verify", revision+"^{commit}")
	return err == nil
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
