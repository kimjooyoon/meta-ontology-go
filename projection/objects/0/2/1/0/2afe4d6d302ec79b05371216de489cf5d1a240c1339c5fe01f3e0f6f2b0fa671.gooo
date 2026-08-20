package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	root := flag.String("root", ".", "repository root")
	storageRoot := flag.String("storage-root", os.Getenv("GOOO_STORAGE_ROOT"), "physical storage root")
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
	if err := run(*root, *storageRoot, *from, *to, *head, *base, *branch, *expectedHead, *capsOnly, *skipCaps); err != nil {
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
