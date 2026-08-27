package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/reproducibilitysemantics"
)

func main() {
	mode := flag.String("mode", "", "produce or judge")
	sourcePath := flag.String("source", "", "Gooo source path")
	headSHA := flag.String("head-sha", "", "exact source commit")
	receiptPath := flag.String("receipt", "", "producer receipt path")
	outputPath := flag.String("output", "", "judgment or receipt output path")
	check := flag.Bool("check", false, "require a discharged independent judgment")
	flag.Parse()
	if *sourcePath == "" || *outputPath == "" {
		fail("-source and -output are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail("read source: %v", err)
	}
	switch *mode {
	case "produce":
		if *headSHA == "" {
			fail("-head-sha is required in produce mode")
		}
		if err := reproducibilitysemantics.WriteJSON(*outputPath,
			reproducibilitysemantics.Produce(*sourcePath, *headSHA, source)); err != nil {
			fail("write receipt: %v", err)
		}
	case "judge":
		if *receiptPath == "" || *headSHA == "" {
			fail("-receipt and -head-sha are required in judge mode")
		}
		receipt, err := reproducibilitysemantics.ReadReceipt(*receiptPath)
		if err != nil {
			fail("%v", err)
		}
		judgment := reproducibilitysemantics.Judge(*sourcePath, *headSHA, source, receipt)
		if *check {
			if err := reproducibilitysemantics.ValidateJudgment(*sourcePath, *headSHA, source, receipt, judgment); err != nil {
				fail("%v", err)
			}
		}
		if err := reproducibilitysemantics.WriteJSON(*outputPath, judgment); err != nil {
			fail("write judgment: %v", err)
		}
		fmt.Printf("reproducibility semantics: decision=%s matrix=%d/%d byte=%d/%d meaning=%d/%d joint=%d/%d counterexamples=%d/%d open=%d/%d source-binding=%d/%d semantic-causality=%d/%d\n",
			judgment.Decision, judgment.Summary.CaseMatrix.Numerator, judgment.Summary.CaseMatrix.Denominator,
			judgment.Summary.ByteClaim.Numerator, judgment.Summary.ByteClaim.Denominator,
			judgment.Summary.MeaningClaim.Numerator, judgment.Summary.MeaningClaim.Denominator,
			judgment.Summary.JointClaim.Numerator, judgment.Summary.JointClaim.Denominator,
			judgment.Summary.Counterexamples.Numerator, judgment.Summary.Counterexamples.Denominator,
			judgment.Summary.OpenCases.Numerator, judgment.Summary.OpenCases.Denominator,
			judgment.Summary.SourceDigestBinding.Numerator, judgment.Summary.SourceDigestBinding.Denominator,
			judgment.Summary.SemanticCausality.Numerator, judgment.Summary.SemanticCausality.Denominator)
	default:
		fail("-mode must be produce or judge")
	}
}

func fail(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", arguments...)
	os.Exit(1)
}
