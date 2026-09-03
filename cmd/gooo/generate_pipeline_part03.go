package main

import (
	"fmt"
	"io"
	"path/filepath"
)

func reportGenerateSuccess(options generateOptions, input generateInput, artifacts generateArtifacts, jsonMode bool, stdout io.Writer) int {
	if !jsonMode {
		fmt.Fprintf(stdout, "generated: %s\n", filepath.Join(options.outputDir, generatedFileName))
		return exitOK
	}
	report := newJSONReport("generate", "ok", options.filename, syntaxCLIDiagnostics(input.diagnostics))
	report.Output = artifacts.output
	report.Manifest = artifacts.manifestPath
	report.PreviousGo = options.previousGo
	report.ProtectedBytesEqual = &artifacts.manifest.ProtectedBytesEqual
	report.SemanticHash = artifacts.ir.StableHash()
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
}

func parseGenerateArguments(args []string) (generateOptions, error) {
	usage := "usage: gooo generate <file.gooo> --out <directory> [--retention-report] [--certificate FILE --contract FILE --observation FILE --proposal FILE --authorization FILE --adoption FILE]"
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
		switch args[index] {
		case "--out":
			if options.outputDir != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.outputDir = value
		case "--previous-go":
			if options.previousGo != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.previousGo = value
		case "--manifest":
			if options.manifestPath != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.manifestPath = value
		case "--certificate", "--retained-certificate":
			if options.retainedCertificateFilename != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.retainedCertificateFilename = value
		case "--contract", "--retention-contract":
			if options.retentionContractFilename != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.retentionContractFilename = value
		case "--observation", "--retention-observation":
			if options.retentionObservationFilename != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.retentionObservationFilename = value
		case "--proposal", "--retention-proposal":
			if options.retentionProposalFilename != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.retentionProposalFilename = value
		case "--authorization", "--retention-authorization":
			if options.retentionAuthorizationFilename != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.retentionAuthorizationFilename = value
		case "--adoption", "--retention-adoption":
			if options.retentionAdoptionFilename != "" {
				return generateOptions{}, fmt.Errorf("%s", usage)
			}
			options.retentionAdoptionFilename = value
		default:
			return generateOptions{}, fmt.Errorf("%s", usage)
		}
		index++
	}
	if options.outputDir == "" {
		return generateOptions{}, fmt.Errorf("%s", usage)
	}
	return options, nil
}
