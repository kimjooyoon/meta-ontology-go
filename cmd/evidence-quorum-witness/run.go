package main

import (
	"fmt"
	"os"
	"strings"

	quorum "github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorum"
)

func run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	if options.check != "" {
		report, err := quorum.LoadReport(options.check)
		if err != nil {
			return err
		}
		return quorum.Validate(report)
	}
	contractRaw, err := os.ReadFile(options.contract)
	if err != nil {
		return err
	}
	contract, err := quorum.DecodeContract(contractRaw)
	if err != nil {
		return err
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		return err
	}
	if options.mode == "emit" {
		return emitReceipt(options, contract, source)
	}
	if options.mode != "evaluate" {
		return fmt.Errorf("unknown mode %q", options.mode)
	}
	receiptGroups, err := readReceiptGroups(options.receipts)
	if err != nil {
		return err
	}
	report := quorum.Evaluate(quorum.Input{Contract: contract, HeadSHA: options.head,
		SourcePath: options.sourcePath, Source: source, CaseReceipts: receiptGroups})
	if err := quorum.WriteReport(options.out, report); err != nil {
		return err
	}
	if err := quorum.Validate(report); err != nil {
		return err
	}
	fmt.Printf("evidence quorum: %s %d/%d groups=%d duplicates=%d conflicts=%d\n",
		report.Decision, report.Summary.CasesSatisfied, report.Summary.CasesTotal,
		report.Summary.IndependentGroupsTotal, report.Summary.DuplicateEvidenceTotal,
		report.Summary.ConflictCases)
	return nil
}

func readReceiptGroups(spec string) ([][][]byte, error) {
	groups := strings.Split(spec, ";")
	result := make([][][]byte, 0, len(groups))
	for _, groupSpec := range groups {
		paths := strings.Split(groupSpec, ",")
		group := make([][]byte, 0, len(paths))
		for _, path := range paths {
			if path == "" {
				continue
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			group = append(group, raw)
		}
		result = append(result, group)
	}
	return result, nil
}

func emitReceipt(options options, contract quorum.Contract, source []byte) error {
	if options.sourcePath == "" {
		options.sourcePath = contract.SourcePath
	}
	claim := contract.Claim
	receipt := quorum.Receipt{
		Schema:        quorum.ReceiptSchema,
		HeadSHA:       options.head,
		SourcePath:    options.sourcePath,
		SourceDigest:  sourceDigest(source),
		Producer:      claim.Producer,
		Consumer:      claim.Consumer,
		MetaOperation: claim.MetaOperation,
		ProofChoice:   claim.ProofChoice,
		Decision:      quorum.DecisionPass,
		Resolution:    quorum.ResolutionExact,
		Evidence: []quorum.Evidence{{
			ID:            options.evidenceID,
			ClaimID:       claim.ID,
			OriginGroup:   options.originGroup,
			Role:          options.role,
			Producer:      claim.Producer,
			Consumer:      claim.Consumer,
			MetaOperation: claim.MetaOperation,
			ProofChoice:   claim.ProofChoice,
			Value:         options.value,
			ConfidenceBPS: options.confidenceBPS,
			SourcePath:    options.sourcePath,
			SourceDigest:  sourceDigest(source),
		}},
		RepositoryWrites:  0,
		MutationAuthority: false,
	}
	return quorum.WriteReceipt(options.out, quorum.SealReceipt(receipt))
}

func sourceDigest(source []byte) string {
	return quorum.SourceDigest(source)
}
