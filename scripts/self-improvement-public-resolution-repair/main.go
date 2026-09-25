package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "", "run or verify")
	source := flag.String("source", "", "canonical partial-reuse .gooo source")
	testContract := flag.String("test-contract", "", "canonical generated project test contract")
	gooo := flag.String("gooo", "", "compiled public gooo command")
	orchestration := flag.String("orchestration-report", "", "closed v14 orchestration report")
	v15Report := flag.String("v15-report", "", "v15 partial-reuse aggregate report")
	v15Hidden := flag.String("v15-hidden", "", "v15 hidden-dependency REFUTED record")
	repoRoot := flag.String("repo-root", "", "repository root")
	out := flag.String("out", "", "caller-owned evidence directory")
	reportPath := flag.String("report", "", "resolution-repair report to verify")
	humanOutput := flag.String("human-output", "", "human verification dossier")
	flag.Parse()

	var err error
	switch *mode {
	case "run":
		err = run(runInput{Source: *source, TestContract: *testContract, Gooo: *gooo, OrchestrationReport: *orchestration, V15Report: *v15Report, V15Hidden: *v15Hidden, RepoRoot: *repoRoot, Out: *out})
	case "verify":
		err = verify(*reportPath, *humanOutput)
	default:
		err = errors.New("usage: self-improvement-public-resolution-repair -mode run|verify")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
