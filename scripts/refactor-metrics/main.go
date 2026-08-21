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
	subject := flag.String("subject", "", "one exact indicator subject to apply")
	check := flag.Bool("check", false, "validate every selected operation without writing")
	flag.Parse()
	if err := run(options{root: *root, metrics: *metrics, sha: *sha, subject: *subject, check: *check}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
