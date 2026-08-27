package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/externaloraclehumilityconsumer"
	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/externaloraclehumilityproducer"
)

type options struct {
	head, source, contract, references, current, mismatchCapsule, absenceCapsule string
	interventionSource, commentSource, effects, independence, output             string
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	opts, err := parseOptions(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	contractRaw, err := read(opts.contract)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	referencesRaw, err := read(opts.references)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	currentRaw, err := read(opts.current)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	effectsRaw, err := readOptional(opts.effects)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	independenceRaw, err := readOptional(opts.independence)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	source, err := read(opts.source)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	receipt, err := producer.ProduceSourceReceipt(opts.head, opts.source, source)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	receiptRaw, err := consumer.Encode(receipt)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	baseInput := consumer.Input{Subject: opts.head, SourcePath: opts.source, Source: source, Contract: contractRaw, Receipt: receiptRaw, Capsule: referencesRaw, Current: currentRaw, Effects: effectsRaw, Independence: independenceRaw}
	agreement, err := consumer.Judge(baseInput)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	mismatchCapsule, err := read(opts.mismatchCapsule)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	absenceCapsule, err := read(opts.absenceCapsule)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	mismatchInput := baseInput
	mismatchInput.Capsule, mismatchInput.Current, mismatchInput.Conformance = mismatchCapsule, nil, true
	mismatch, err := consumer.Judge(mismatchInput)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	absenceInput := baseInput
	absenceInput.Capsule, absenceInput.Current, absenceInput.Conformance = absenceCapsule, nil, true
	absence, err := consumer.Judge(absenceInput)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	interventionSource, err := read(opts.interventionSource)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	interventionReceipt, err := producer.ProduceSourceReceipt(opts.head, opts.interventionSource, interventionSource)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	interventionReceiptRaw, err := consumer.Encode(interventionReceipt)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	interventionInput := baseInput
	interventionInput.SourcePath, interventionInput.Source, interventionInput.Receipt = opts.interventionSource, interventionSource, interventionReceiptRaw
	intervention, err := consumer.Judge(interventionInput)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	commentSource, err := read(opts.commentSource)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	commentReceipt, err := producer.ProduceSourceReceipt(opts.head, opts.commentSource, commentSource)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	commentReceiptRaw, err := consumer.Encode(commentReceipt)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	commentInput := baseInput
	commentInput.SourcePath, commentInput.Source, commentInput.Receipt = opts.commentSource, commentSource, commentReceiptRaw
	comment, err := consumer.Judge(commentInput)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	agreement = consumer.FinalizeCausality(agreement, intervention, comment)
	suite, err := consumer.BuildSuite(contractRaw, opts.head, agreement, mismatch, absence, intervention, comment)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	values := map[string]any{
		"source-receipt.json":      receipt,
		"agreement-report.json":    agreement,
		"mismatch-report.json":     mismatch,
		"absence-report.json":      absence,
		"intervention-report.json": intervention,
		"comment-report.json":      comment,
		"suite.json":               suite,
	}
	if err := os.MkdirAll(opts.output, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	for name, value := range values {
		raw, encodeErr := consumer.Encode(value)
		if encodeErr != nil {
			fmt.Fprintln(stderr, encodeErr)
			return 2
		}
		if writeErr := os.WriteFile(filepath.Join(opts.output, name), raw, 0o644); writeErr != nil {
			fmt.Fprintln(stderr, writeErr)
			return 2
		}
	}
	fmt.Fprintf(stdout, "external oracle humility: %s %d/%d cases=%d/%d agreement=%s/%s authority=%s current-bytes=%d/%d semantic-extraction=%d/3 semantic-agreement=%d/3\n", suite.Decision, agreement.Completed, agreement.Total, suite.CasesSatisfied, suite.CasesTotal, agreement.ReferenceAgreement, agreement.SemanticAgreement, agreement.SemanticAuthority, agreement.CurrentByteObservations, agreement.CurrentReferenceTotal, agreement.FixedDenominator.SemanticExtraction.Completed, agreement.FixedDenominator.SemanticAgreement.Completed)
	if suite.Decision != "HUMILITY_MODEL_BOUND" || suite.Resolution != "EXACT" {
		return 1
	}
	return 0
}

func read(path string) ([]byte, error) { return os.ReadFile(path) }

func readOptional(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	return read(path)
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var result options
	flags := flag.NewFlagSet("external-oracle-humility-witness", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&result.head, "head", "", "exact subject SHA")
	flags.StringVar(&result.source, "source", "", "Gooo source path")
	flags.StringVar(&result.contract, "contract", "", "humility contract path")
	flags.StringVar(&result.references, "references", "", "historical reference capsule path")
	flags.StringVar(&result.current, "current", "", "Actions current observation path")
	flags.StringVar(&result.mismatchCapsule, "mismatch-capsule", "", "historical mismatch capsule path")
	flags.StringVar(&result.absenceCapsule, "absence-capsule", "", "historical absence capsule path")
	flags.StringVar(&result.interventionSource, "intervention-source", "", "semantic intervention source path")
	flags.StringVar(&result.commentSource, "comment-source", "", "comment-only source path")
	flags.StringVar(&result.effects, "effects", "", "repository effects snapshot path")
	flags.StringVar(&result.independence, "independence", "", "producer/consumer dependency snapshot path")
	flags.StringVar(&result.output, "output", "", "output directory")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if result.head == "" || result.source == "" || result.contract == "" || result.references == "" || result.current == "" || result.mismatchCapsule == "" || result.absenceCapsule == "" || result.interventionSource == "" || result.commentSource == "" || result.output == "" {
		return options{}, fmt.Errorf("source, capsule, observation, intervention, and output flags are required")
	}
	return result, nil
}
