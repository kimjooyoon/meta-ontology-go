package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/reproducibilitysemantics"
	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/reproducibilitysemanticsconsumer"
)

func main() {
	mode := flag.String("mode", "", "produce, judge, or intervention")
	sourcePath := flag.String("source", "", "Gooo source path")
	semanticSourcePath := flag.String("semantic-source", "", "semantic intervention source path")
	presentationSourcePath := flag.String("presentation-source", "", "presentation intervention source path")
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
		receipt, err := consumer.ReadReceipt(*receiptPath)
		if err != nil {
			fail("%v", err)
		}
		judgment := consumer.Judge(*sourcePath, *headSHA, source, receipt)
		if *check {
			if err := consumer.ValidateJudgment(*sourcePath, *headSHA, source, receipt, judgment); err != nil {
				fail("%v", err)
			}
		}
		if err := consumer.WriteJSON(*outputPath, judgment); err != nil {
			fail("write judgment: %v", err)
		}
		fmt.Printf("reproducibility semantics: conformance=%s/%s subject=%s/%s matrix=%d/%d byte=%d/%d meaning=%d/%d joint=%d/%d counterexamples=%d/%d open=%d/%d source-binding=%d/%d semantic-causality=%d/%d\n",
			judgment.ConformanceDecision, judgment.ConformanceResolution, judgment.SubjectDecision, judgment.SubjectResolution,
			judgment.Summary.CaseMatrix.Numerator, judgment.Summary.CaseMatrix.Denominator,
			judgment.Summary.ByteClaim.Numerator, judgment.Summary.ByteClaim.Denominator,
			judgment.Summary.MeaningClaim.Numerator, judgment.Summary.MeaningClaim.Denominator,
			judgment.Summary.JointClaim.Numerator, judgment.Summary.JointClaim.Denominator,
			judgment.Summary.Counterexamples.Numerator, judgment.Summary.Counterexamples.Denominator,
			judgment.Summary.OpenCases.Numerator, judgment.Summary.OpenCases.Denominator,
			judgment.Summary.SourceDigestBinding.Numerator, judgment.Summary.SourceDigestBinding.Denominator,
			judgment.Summary.SemanticCausality.Numerator, judgment.Summary.SemanticCausality.Denominator)
	case "intervention":
		if *headSHA == "" || *semanticSourcePath == "" || *presentationSourcePath == "" {
			fail("-head-sha, -semantic-source, and -presentation-source are required in intervention mode")
		}
		semanticSource, err := os.ReadFile(*semanticSourcePath)
		if err != nil {
			fail("read semantic source: %v", err)
		}
		presentationSource, err := os.ReadFile(*presentationSourcePath)
		if err != nil {
			fail("read presentation source: %v", err)
		}
		baseJudgment := judgeProduced(*sourcePath, *headSHA, source)
		semanticJudgment := judgeProduced(*semanticSourcePath, *headSHA, semanticSource)
		presentationJudgment := judgeProduced(*presentationSourcePath, *headSHA, presentationSource)
		artifact, err := consumer.BuildInterventionArtifact(baseJudgment, semanticJudgment, presentationJudgment)
		if err != nil {
			fail("build intervention artifact: %v", err)
		}
		if err := consumer.ValidateIntervention(artifact); err != nil {
			fail("validate intervention artifact: %v", err)
		}
		if err := consumer.WriteJSON(*outputPath, artifact); err != nil {
			fail("write intervention artifact: %v", err)
		}
		fmt.Printf("reproducibility semantics intervention: denominator=%d decision=%s resolution=%s\n", artifact.Denominator, artifact.Decision, artifact.Resolution)
	default:
		fail("-mode must be produce, judge, or intervention")
	}
}

func judgeProduced(sourcePath, headSHA string, source []byte) consumer.Judgment {
	receipt := reproducibilitysemantics.Produce(sourcePath, headSHA, source)
	raw, err := json.Marshal(receipt)
	if err != nil {
		fail("encode intervention receipt: %v", err)
	}
	judgment := consumer.Judge(sourcePath, headSHA, source, raw)
	if err := consumer.ValidateJudgment(sourcePath, headSHA, source, raw, judgment); err != nil {
		fail("intervention judgment: %v", err)
	}
	return judgment
}

func fail(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", arguments...)
	os.Exit(1)
}
