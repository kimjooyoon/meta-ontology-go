package main

import (
	"fmt"
	"os"
	"strings"

	consumer "github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumconsumer"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumpolicy"
)

func run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	if options.check != "" {
		report, err := consumer.LoadReport(options.check)
		if err != nil {
			return err
		}
		return consumer.Validate(report)
	}
	if options.policy == "" || options.source == "" || options.head == "" || options.sourcePath == "" || options.cases == "" {
		return fmt.Errorf("--policy, --source, --head, --source-path, and --cases are required")
	}
	policyRaw, err := os.ReadFile(options.policy)
	if err != nil {
		return err
	}
	policy, err := evidencequorumpolicy.Parse(options.policy, policyRaw)
	if err != nil {
		return err
	}
	if err := consumer.ValidatePolicy(policy); err != nil {
		return err
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		return err
	}
	cases, err := readCases(options.cases)
	if err != nil {
		return err
	}
	report := consumer.Evaluate(consumer.Input{Policy: policy, HeadSHA: options.head, SourcePath: options.sourcePath, Source: source, Cases: cases})
	if err := consumer.WriteReport(options.out, report); err != nil {
		return err
	}
	if err := consumer.Validate(report); err != nil {
		return err
	}
	fmt.Printf("evidence quorum: %s cases=%d/%d current=%d synthetic=%d groups=%d collapsed=%d\n",
		report.Decision, report.Summary.CasesSatisfied, report.Summary.CasesTotal,
		report.Summary.CurrentEvidenceTotal, report.Summary.SyntheticEvidenceTotal,
		report.Summary.DistinctProvenanceGroups, report.Summary.CollapsedReplicas)
	return nil
}

func readCases(spec string) ([]consumer.CaseInput, error) {
	var result []consumer.CaseInput
	for _, caseSpec := range strings.Split(spec, ";") {
		id, paths, ok := strings.Cut(caseSpec, "=")
		if !ok || strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("invalid case spec %q", caseSpec)
		}
		var receipts [][]byte
		for _, path := range strings.Split(paths, ",") {
			if path = strings.TrimSpace(path); path != "" {
				raw, err := os.ReadFile(path)
				if err != nil {
					return nil, err
				}
				receipts = append(receipts, raw)
			}
		}
		result = append(result, consumer.CaseInput{ID: strings.TrimSpace(id), Receipts: receipts})
	}
	return result, nil
}
