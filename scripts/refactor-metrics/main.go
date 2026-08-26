package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	root := flag.String("root", ".", "repository root")
	metrics := flag.String("metrics", "", "exact-SHA source metrics JSON")
	sha := flag.String("sha", "", "expected metrics commit SHA")
	baseSHA := flag.String("base-sha", "", "generation base commit SHA")
	plan := flag.String("plan", "", "write a canonical self-improvement plan")
	subject := flag.String("subject", "", "one exact indicator subject to apply")
	check := flag.Bool("check", false, "validate every selected operation without writing")
	flag.Parse()
	if err := run(options{root: *root, metrics: *metrics, sha: *sha, baseSHA: *baseSHA, plan: *plan, subject: *subject, check: *check}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
