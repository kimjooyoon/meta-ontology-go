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
	profile                          string
	profilePackage                   string
	profileNamespace                 string
	profileProjectRoot               string
	filename                         string
	outputDir                        string
	previousGo                       string
	manifestPath                     string
	retentionReport                  bool
	retainedCertificateFilename      string
	continuityCertificateFilename    string
	compatibilityCertificateFilename string
	retentionContractFilename        string
	retentionObservationFilename     string
	retentionProposalFilename        string
	retentionAuthorizationFilename   string
	retentionAdoptionFilename        string
	observationLedgerDir             string
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
	if options.profile == "" {
		if options.profilePackage != "" || options.profileNamespace != "" || options.profileProjectRoot != "" {
			return generateOptions{}, fmt.Errorf("%s", usage)
		}
	} else {
		if options.profile != "meta-policy-compilation-v3" || options.profilePackage == "" || options.profileNamespace == "" || options.profileProjectRoot == "" {
			return generateOptions{}, fmt.Errorf("%s", usage)
		}
		if options.previousGo != "" || options.manifestPath != "" || options.retentionReport || options.publicRetentionRequested() || options.continuityCertificateFilename != "" || options.compatibilityCertificateFilename != "" || options.observationLedgerDir != "" {
			return generateOptions{}, fmt.Errorf("%s", usage)
		}
	}
	if options.observationLedgerDir != "" && (options.retentionReport || options.publicRetentionRequested()) {
		return generateOptions{}, fmt.Errorf("%s", usage)
	}
	if options.continuityCertificateFilename != "" && (options.observationLedgerDir != "" || options.retentionReport || options.publicRetentionRequested() || options.compatibilityCertificateFilename != "") {
		return generateOptions{}, fmt.Errorf("%s", usage)
	}
	if options.compatibilityCertificateFilename != "" && (options.observationLedgerDir != "" || options.retentionReport || options.publicRetentionRequested()) {
		return generateOptions{}, fmt.Errorf("%s", usage)
	}
	return options, nil
}

func setGenerateOption(options *generateOptions, name, value string) bool {
	return setProfileGenerateOption(options, name, value) ||
		setOutputGenerateOption(options, name, value) ||
		setCertificateGenerateOption(options, name, value) ||
		setRetentionGenerateOption(options, name, value)
}

func setProfileGenerateOption(options *generateOptions, name, value string) bool {
	switch name {
	case "--profile":
		return setGenerateString(&options.profile, value)
	case "--profile-package":
		return setGenerateString(&options.profilePackage, value)
	case "--profile-namespace":
		return setGenerateString(&options.profileNamespace, value)
	case "--profile-project-root":
		return setGenerateString(&options.profileProjectRoot, value)
	default:
		return false
	}
}

func setOutputGenerateOption(options *generateOptions, name, value string) bool {
	switch name {
	case "--out":
		return setGenerateString(&options.outputDir, value)
	case "--previous-go":
		return setGenerateString(&options.previousGo, value)
	case "--manifest":
		return setGenerateString(&options.manifestPath, value)
	default:
		return false
	}
}

func setCertificateGenerateOption(options *generateOptions, name, value string) bool {
	switch name {
	case "--certificate", "--retained-certificate":
		return setGenerateString(&options.retainedCertificateFilename, value)
	case "--continuity-certificate":
		return setGenerateString(&options.continuityCertificateFilename, value)
	case "--compatibility-certificate":
		return setGenerateString(&options.compatibilityCertificateFilename, value)
	default:
		return false
	}
}

func setRetentionGenerateOption(options *generateOptions, name, value string) bool {
	switch name {
	case "--contract", "--retention-contract":
		return setGenerateString(&options.retentionContractFilename, value)
	case "--observation", "--retention-observation":
		return setGenerateString(&options.retentionObservationFilename, value)
	case "--proposal", "--retention-proposal":
		return setGenerateString(&options.retentionProposalFilename, value)
	case "--authorization", "--retention-authorization":
		return setGenerateString(&options.retentionAuthorizationFilename, value)
	case "--adoption", "--retention-adoption":
		return setGenerateString(&options.retentionAdoptionFilename, value)
	case "--observation-ledger", "--observation-output":
		return setGenerateString(&options.observationLedgerDir, value)
	default:
		return false
	}
}

func setGenerateString(target *string, value string) bool {
	if *target != "" {
		return false
	}
	*target = value
	return true
}
