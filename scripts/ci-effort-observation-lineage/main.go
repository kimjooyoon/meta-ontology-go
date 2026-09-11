package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "run", "run, observe, or verify")
	source := flag.String("source", "examples/ci-effort-observation/main.gooo", "canonical workflow-lineage source")
	out := flag.String("out", "", "conformance output directory")
	trigger := flag.String("trigger", "", "exact trigger metadata")
	run := flag.String("run", "", "exact source workflow run metadata")
	artifacts := flag.String("artifacts", "", "exact source artifact metadata")
	reportPath := flag.String("report", "", "verification report path")
	humanPath := flag.String("human-output", "", "verification human output path")
	flag.Parse()

	var err error
	switch *mode {
	case "run":
		err = runConformance(*source, *out)
	case "observe":
		err = runObserve(runInput{Source: *source, Out: *out, Trigger: *trigger, Run: *run, Artifacts: *artifacts})
	case "verify":
		err = verifyReport(*reportPath, *humanPath)
	default:
		err = fmt.Errorf("unsupported mode %q", *mode)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
