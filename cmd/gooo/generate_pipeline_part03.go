package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
)

func reportGenerateSuccess(options generateOptions, input generateInput, artifacts generateArtifacts, discovery *publicdiscovery.Result, jsonMode bool, stdout io.Writer) int {
	if !jsonMode {
		fmt.Fprintf(stdout, "generated: %s\n", filepath.Join(options.outputDir, generatedFileName))
		if discovery != nil {
			fmt.Fprintf(stdout, "observation: %s (%s)\n", discovery.Report.MachineReportPath, discovery.Report.Decision)
			if discovery.Report.CandidatesEmitted > 0 {
				fmt.Fprintf(stdout, "candidate: %s\n", discovery.CandidatePath)
			}
		}
		return exitOK
	}
	report := newJSONReport("generate", "ok", options.filename, syntaxCLIDiagnostics(input.diagnostics))
	report.Output = artifacts.output
	report.Manifest = artifacts.manifestPath
	report.PreviousGo = options.previousGo
	report.ProtectedBytesEqual = &artifacts.manifest.ProtectedBytesEqual
	report.SemanticHash = artifacts.ir.StableHash()
	if discovery != nil {
		report.ObservationReport = discovery.Report.MachineReportPath
		report.ObservationDecision = discovery.Report.Decision
		report.ObservationReason = discovery.Report.Reason
		if discovery.Report.CandidatesEmitted > 0 {
			report.ObservationCandidate = discovery.CandidatePath
		}
	}
	if err := writeJSONReport(stdout, report); err != nil {
		return exitFailure
	}
	return exitOK
}

const generatedManifestFileName = "semantic.gooo.manifest.jsonl"

type generateOptions struct {
	filename                       string
	outputDir                      string
	previousGo                     string
	manifestPath                   string
	retentionReport                bool
	retainedCertificateFilename    string
	retentionContractFilename      string
	retentionObservationFilename   string
	retentionProposalFilename      string
	retentionAuthorizationFilename string
	retentionAdoptionFilename      string
	observationLedgerDir           string
}

func parseGenerateArguments(args []string) (generateOptions, error) {
	usage := "usage: gooo generate <file.gooo> --out <directory>"
	if len(args) == 0 {
		return generateOptions{}, fmt.Errorf("%s", usage)
	}
	options := generateOptions{filename: args[0]}
	for index := 1; index < len(args); index++ {
		if args[index] == "--retention-report" {
			if options.retentionReport {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.retentionReport = true
			continue
		}
		if index+1 >= len(args) {
			return generateOptions{}, fmt.Errorf("%s", usage)
		}
		value := args[index+1]
		if value == "" {
			return generateOptions{}, fmt.Errorf("%s", usage)
		}
		if !setGenerateOption(&options, args[index], value) {
			return generateOptions{}, fmt.Errorf("%s", usage)
		}
		index++
	}
	if options.outputDir == "" {
		return generateOptions{}, fmt.Errorf("%s", usage)
	}
	if options.observationLedgerDir != "" && (options.retentionReport || options.publicRetentionRequested()) {
		return generateOptions{}, fmt.Errorf("%s", usage)
	}
	return options, nil
}

func setGenerateOption(options *generateOptions, name, value string) bool {
	switch name {
	case "--out":
		if options.outputDir != "" {
			return false
		}
		options.outputDir = value
	case "--previous-go":
		if options.previousGo != "" {
			return false
		}
		options.previousGo = value
	case "--manifest":
		if options.manifestPath != "" {
			return false
		}
		options.manifestPath = value
	case "--certificate", "--retained-certificate":
		if options.retainedCertificateFilename != "" {
			return false
		}
		options.retainedCertificateFilename = value
	case "--contract", "--retention-contract":
		if options.retentionContractFilename != "" {
			return false
		}
		options.retentionContractFilename = value
	case "--observation", "--retention-observation":
		if options.retentionObservationFilename != "" {
			return false
		}
		options.retentionObservationFilename = value
	case "--proposal", "--retention-proposal":
		if options.retentionProposalFilename != "" {
			return false
		}
		options.retentionProposalFilename = value
	case "--authorization", "--retention-authorization":
		if options.retentionAuthorizationFilename != "" {
			return false
		}
		options.retentionAuthorizationFilename = value
	case "--adoption", "--retention-adoption":
		if options.retentionAdoptionFilename != "" {
			return false
		}
		options.retentionAdoptionFilename = value
	case "--observation-ledger", "--observation-output":
		if options.observationLedgerDir != "" {
			return false
		}
		options.observationLedgerDir = value
	default:
		return false
	}
	return true
}
