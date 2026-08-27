package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "produce", "produce, judge, receipt-only, or artifact")
	source := flag.String("source", "", "raw .gooo source path")
	receipt := flag.String("receipt", "", "receipt path")
	output := flag.String("output", "", "output JSON path")
	root := flag.String("repo-root", "", "repository root for CI status observation")
	before := flag.String("before-state", "", "before status snapshot output/input")
	after := flag.String("after-state", "", "after status snapshot output/input")
	expect := flag.String("expect", "", "expected decision")
	baseline := flag.String("replay-baseline", "", "canonical artifact used as the replay baseline")
	flag.Parse()
	if err := run(*mode, *source, *receipt, *output, *root, *before, *after, *expect, *baseline); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
