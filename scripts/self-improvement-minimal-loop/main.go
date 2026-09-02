package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementloop"
)

func main() {
	graphPath := flag.String("graph", "", "released semantic graph JSON")
	casePath := flag.String("case", "", "loop case JSON")
	outputDir := flag.String("output", "", "caller-owned temporary output directory")
	expected := flag.String("expected", "", "optional expected decision")
	flag.Parse()
	if *graphPath == "" || *casePath == "" || *outputDir == "" {
		fail("-graph, -case, and -output are required")
	}
	graphData, err := os.ReadFile(*graphPath)
	if err != nil {
		fail("read graph: %v", err)
	}
	caseData, err := os.ReadFile(*casePath)
	if err != nil {
		fail("read case: %v", err)
	}
	graph, err := selfimprovementloop.DecodeGraph(graphData)
	if err != nil {
		fail("%v", err)
	}
	input, err := selfimprovementloop.DecodeInput(caseData)
	if err != nil {
		fail("%v", err)
	}
	artifacts, err := selfimprovementloop.Run(graph, input)
	if err != nil {
		fail("evaluate loop: %v", err)
	}
	if err := selfimprovementloop.WriteArtifacts(*outputDir, artifacts); err != nil {
		fail("write caller-owned artifacts: %v", err)
	}
	if *expected != "" && artifacts.Report.Decision != *expected {
		fail("decision = %s, want %s", artifacts.Report.Decision, *expected)
	}
	fmt.Printf("scenario=%s decision=%s cells=%d pair_matched=%t\n", artifacts.Report.Scenario, artifacts.Report.Decision, len(artifacts.Report.Cells), artifacts.Report.PairMatched)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(2)
}
