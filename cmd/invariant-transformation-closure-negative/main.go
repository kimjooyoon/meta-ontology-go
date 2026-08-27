package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/closureverifier"
)

func main() {
	mode := flag.String("mode", "", "negative fixture mode: evidence-reseal, comment-reseal, comment-semantic-digest, comment-gate-false, or semantic-case-reseal")
	input := flag.String("input", "", "input closure receipt for evidence-reseal")
	intervention := flag.String("intervention-report", "", "producer intervention report for comment-reseal")
	consumer := flag.String("intervention-consumer-receipt", "", "independent intervention consumer receipt for comment-reseal")
	output := flag.String("output", "", "tampered closure output")
	outputIntervention := flag.String("output-intervention-report", "", "tampered intervention report output")
	outputConsumer := flag.String("output-consumer-receipt", "", "tampered consumer receipt output")
	flag.Parse()

	switch *mode {
	case "evidence-reseal":
		raw, err := os.ReadFile(*input)
		if err != nil {
			fail(err.Error())
		}
		result, err := closureverifier.ResealEvidenceDigestFixture(raw)
		if err != nil {
			fail(err.Error())
		}
		write(*output, result)
	case "comment-reseal":
		interventionRaw, err := os.ReadFile(*intervention)
		if err != nil {
			fail(err.Error())
		}
		consumerRaw, err := os.ReadFile(*consumer)
		if err != nil {
			fail(err.Error())
		}
		mutatedIntervention, mutatedConsumer, err := closureverifier.ResealCommentFixture(interventionRaw, consumerRaw)
		if err != nil {
			fail(err.Error())
		}
		write(*outputIntervention, mutatedIntervention)
		write(*outputConsumer, mutatedConsumer)
	case "comment-semantic-digest", "comment-gate-false":
		interventionRaw, err := os.ReadFile(*intervention)
		if err != nil {
			fail(err.Error())
		}
		consumerRaw, err := os.ReadFile(*consumer)
		if err != nil {
			fail(err.Error())
		}
		var mutatedIntervention, mutatedConsumer []byte
		if *mode == "comment-semantic-digest" {
			mutatedIntervention, mutatedConsumer, err = closureverifier.ResealCommentSemanticDigestFixture(interventionRaw, consumerRaw)
		} else {
			mutatedIntervention, mutatedConsumer, err = closureverifier.ResealCommentGateFixture(interventionRaw, consumerRaw)
		}
		if err != nil {
			fail(err.Error())
		}
		write(*outputIntervention, mutatedIntervention)
		write(*outputConsumer, mutatedConsumer)
	case "semantic-case-reseal":
		interventionRaw, err := os.ReadFile(*intervention)
		if err != nil {
			fail(err.Error())
		}
		mutated, err := closureverifier.ResealSemanticCaseFixture(interventionRaw)
		if err != nil {
			fail(err.Error())
		}
		write(*outputIntervention, mutated)
	default:
		fail("-mode must be evidence-reseal, comment-reseal, comment-semantic-digest, comment-gate-false, or semantic-case-reseal")
	}
}

func write(path string, raw []byte) {
	if path == "" {
		fail("output path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
