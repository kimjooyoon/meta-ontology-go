package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/hygienicoriginidentity/consumer"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/hygienicoriginidentity/producer"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/hygienicoriginidentity/verify"
)

const baselineSource = "examples/language-hygienic-origin-identity/main.gooo"

func main() {
	source := flag.String("source", baselineSource, "Gooo experiment source")
	head := flag.String("head-sha", "", "exact source commit")
	output := flag.String("output", "", "receipt output outside the repository")
	before := flag.String("before-snapshot", "", "CI git-status snapshot before evaluation")
	after := flag.String("after-snapshot", "", "CI git-status snapshot after evaluation")
	expectation := flag.String("expectation", verify.ExpectationPass, "pass, unknown, or refuted")
	compare := flag.String("compare-source", "", "source used for comment-only invariance")
	check := flag.Bool("check", false, "validate the fixed experiment contract")
	flag.Parse()
	if err := run(*source, *head, *output, *before, *after, *expectation, *compare, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(source, head, output, beforePath, afterPath, expectation, compare string, check bool) error {
	files := os.DirFS(".")
	snapshots, err := readSnapshots(beforePath, afterPath)
	if err != nil {
		return err
	}
	report, err := evaluatePair(files, source, head, snapshots)
	if err != nil {
		return err
	}
	if check {
		switch expectation {
		case verify.ExpectationPass, verify.ExpectationUnknown:
			if err := verify.Validate(files, report, expectation, head); err != nil {
				return err
			}
		case "refuted":
			baseline, err := evaluatePair(files, baselineSource, head, snapshots)
			if err != nil {
				return fmt.Errorf("evaluate baseline: %w", err)
			}
			if err := verify.ValidateIntervention(files, report, baseline, head); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported expectation %q", expectation)
		}
	}
	if compare != "" {
		variant, err := evaluatePair(files, compare, head, snapshots)
		if err != nil {
			return fmt.Errorf("evaluate comparison source: %w", err)
		}
		if report.Source.RawDigest == variant.Source.RawDigest {
			return fmt.Errorf("comment-only fixture did not change raw digest")
		}
		if report.Source.SemanticDigest != variant.Source.SemanticDigest || consumer.ContentDigest(report) != consumer.ContentDigest(variant) || report.Decision != variant.Decision || report.Resolution != variant.Resolution {
			return fmt.Errorf("comment-only fixture changed semantic result")
		}
	}
	if output != "" {
		encoded, err := consumer.Encode(report)
		if err != nil {
			return fmt.Errorf("encode receipt: %w", err)
		}
		if err := os.WriteFile(output, encoded, 0o644); err != nil {
			return fmt.Errorf("write receipt: %w", err)
		}
	}
	fmt.Printf("hygienic origin identity: %s %d/%d classified, target %d/%d preserved, %d open, %d unknown\n", report.Decision, report.Metrics.ClassifiedClaimTotal, report.Metrics.FixedClaimDenominator, report.Metrics.TargetPreservationDischarged, report.Metrics.TargetPreservationExpected, report.Metrics.OpenClaimTotal, report.Metrics.UnknownPathTotal)
	return nil
}

func evaluatePair(files fs.FS, source, head string, snapshots consumer.SnapshotPair) (consumer.Report, error) {
	producerReport, err := producer.Evaluate(files, source)
	if err != nil {
		return consumer.Report{}, err
	}
	consumerReport, err := consumer.Evaluate(files, source, head, snapshots)
	if err != nil {
		return consumer.Report{}, err
	}
	if err := compareProducerReport(producerReport, consumerReport); err != nil {
		return consumer.Report{}, err
	}
	return consumerReport, nil
}

func compareProducerReport(producerReport producer.Report, consumerReport consumer.Report) error {
	if producerReport.RawDigest != consumerReport.Source.RawDigest || producerReport.SemanticDigest != consumerReport.Source.SemanticDigest || len(producerReport.Records) != len(consumerReport.Cases) {
		return fmt.Errorf("producer and consumer canonical digests or case counts diverged")
	}
	for index, record := range producerReport.Records {
		item := consumerReport.Cases[index]
		if record.CaseID != item.ID || record.Spelling != item.Spelling || record.OriginIdentity != item.OriginIdentity || record.DefinitionScope != item.DefinitionScope || record.UseScope != item.UseScope || record.ResolvedUseScope != item.ResolvedUseScope || record.ResolvedIdentity != item.ResolvedIdentity || record.Captured != item.Captured {
			return fmt.Errorf("producer/consumer record %q diverged", record.CaseID)
		}
	}
	return nil
}

func readSnapshots(beforePath, afterPath string) (consumer.SnapshotPair, error) {
	var snapshots consumer.SnapshotPair
	if beforePath == "" && afterPath == "" {
		return snapshots, nil
	}
	if beforePath == "" || afterPath == "" {
		return consumer.SnapshotPair{}, fmt.Errorf("before and after snapshots must be supplied together")
	}
	var err error
	snapshots.Before, err = os.ReadFile(beforePath)
	if err != nil {
		return consumer.SnapshotPair{}, fmt.Errorf("read before snapshot: %w", err)
	}
	snapshots.After, err = os.ReadFile(afterPath)
	if err != nil {
		return consumer.SnapshotPair{}, fmt.Errorf("read after snapshot: %w", err)
	}
	return snapshots, nil
}
