package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependencyjudge"
)

func main() {
	sourcePath := flag.String("source", "", "claim dependency Gooo source")
	caseName := flag.String("case", "", "direct-unknown, refuted, or recovered")
	outputPath := flag.String("output", "", "receipt output path")
	check := flag.Bool("check", false, "run the independent judge after producing the receipt")
	flag.Parse()
	if *sourcePath == "" || *caseName == "" || *outputPath == "" {
		fail("-source, -case, and -output are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	receipt, err := claimdependency.Evaluate(source, *sourcePath, *caseName)
	if err != nil {
		fail(err.Error())
	}
	if *check {
		if _, err := claimdependencyjudge.Judge(receipt); err != nil {
			fail(err.Error())
		}
	}
	writeJSON(*outputPath, receipt)
	fmt.Printf("claim dependency case=%s claims=%d/%d edges=%d direct_unknown=%d blocked=%d refuted=%d recovered=%d depth=%d decision=%s read_only=true repository_writes=0\n", receipt.Subject.Case, receipt.Metrics.ClassifiedClaimTotal, receipt.Metrics.FixedClaimTotal, receipt.Metrics.FixedEdgeTotal, receipt.Metrics.DirectUnknownClaimTotal, receipt.Metrics.DependencyBlockedClaimTotal, receipt.Metrics.RefutedClaimTotal, receipt.Metrics.DependencyRecoveredTotal, receipt.Metrics.MaximumCausePathDepth, receipt.Decision.Value)
}

func writeJSON(path string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
