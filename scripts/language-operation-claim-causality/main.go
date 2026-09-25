package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valuecatalog/causality"
)

func main() {
	inputPath := flag.String("input", "", "operation catalog v3 report")
	discoverRoots := flag.String("discover-roots", "", "comma-separated roots containing operation catalog reports")
	mode := flag.String("mode", "", "success or unknown")
	outputPath := flag.String("output", "", "causality receipt output")
	check := flag.Bool("check", false, "validate the generated receipt")
	flag.Parse()

	if *mode != causality.ModeSuccess && *mode != causality.ModeUnknown {
		fail("-mode must be success or unknown")
	}
	if *outputPath == "" {
		fail("-output is required")
	}
	if (*inputPath == "") == (*discoverRoots == "") {
		fail("exactly one of -input or -discover-roots is required")
	}

	var (
		input  []byte
		source string
		err    error
	)
	if *inputPath != "" {
		input, err = os.ReadFile(*inputPath)
		source = *inputPath
	} else {
		input, source, err = causality.DiscoverReport(strings.Split(*discoverRoots, ","), *mode)
	}
	if err != nil {
		fail(err.Error())
	}
	receipt, err := causality.Evaluate(input, *mode)
	if err != nil {
		fail(err.Error())
	}
	if *check {
		if err := causality.Validate(receipt); err != nil {
			fail(err.Error())
		}
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outputPath, data, 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("claim causality source=%s claims=%d/%d graph=%d/%d direct=%d blocked=%d blocking_edges=%d depth=%d decision=%s semantic_authority=0\n", source, receipt.Metrics.ClassifiedClaimTotal, receipt.Metrics.ContractClaimTotal, receipt.Graph.NodeTotal, receipt.Graph.EdgeTotal, receipt.Metrics.DirectMissingClaimTotal, receipt.Metrics.DependencyBlockedClaimTotal, receipt.Metrics.ObservedBlockingEdgeTotal, receipt.Metrics.MaximumCausePathDepth, receipt.Decision.Value)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
