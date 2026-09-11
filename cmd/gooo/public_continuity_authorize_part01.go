package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publiccontinuity"
)

const continuityAuthorizeUsage = "usage: gooo authorize-discovery CANDIDATE --decision accept|reject --out DIRECTORY"

type continuityAuthorizeOptions struct {
	candidateFilename string
	decision          string
	outputDir         string
}

func runContinuityAuthorize(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseContinuityAuthorizeArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	started := time.Now()
	candidateData, err := reader.ReadFile(options.candidateFilename)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery authorization: read candidate: %v\n", err)
		return exitFailure
	}
	candidate, err := publiccontinuity.DecodeCandidate(candidateData)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery authorization: %v\n", err)
		return exitFailure
	}
	candidateDigest := cache.HashBytes(candidateData).String()
	decision := strings.ToUpper(options.decision)
	reason := publiccontinuity.ReasonAccepted
	if decision == publiccontinuity.DecisionReject {
		reason = publiccontinuity.ReasonRejected
	}
	receipt := publiccontinuity.DecisionReceipt{
		Schema: publiccontinuity.DecisionReceiptSchema, Operation: publiccontinuity.Operation,
		Decision: decision, Reason: reason, ExplicitHumanDecision: true,
		ExecutionAllowed: false, ManualTransformations: 0,
		Binding: publiccontinuity.BindingFromCandidate(candidate, candidateDigest), Candidate: candidate,
		RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0,
	}
	receipt.ReceiptID, err = publiccontinuity.DecisionReceiptContentDigest(receipt)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery authorization: %v\n", err)
		return exitFailure
	}
	if err := publiccontinuity.ValidateDecisionReceipt(receipt); err != nil {
		fmt.Fprintf(stderr, "gooo: discovery authorization: validate receipt: %v\n", err)
		return exitFailure
	}
	receiptData, err := marshalContinuityJSON(receipt)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery authorization: encode receipt: %v\n", err)
		return exitFailure
	}
	if err := writeContinuityArtifacts(options.outputDir, []continuityArtifact{{name: "decision-receipt.json", data: receiptData}}); err != nil {
		fmt.Fprintf(stderr, "gooo: discovery authorization: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "decision receipt: %s (%s, %dms)\n", filepath.Join(options.outputDir, "decision-receipt.json"), decision, continuityWallMS(started))
	return exitOK
}

func parseContinuityAuthorizeArguments(args []string) (continuityAuthorizeOptions, error) {
	if len(args) == 0 {
		return continuityAuthorizeOptions{}, fmt.Errorf("%s", continuityAuthorizeUsage)
	}
	options := continuityAuthorizeOptions{candidateFilename: args[0]}
	for index := 1; index < len(args); index++ {
		if index+1 >= len(args) || args[index+1] == "" {
			return continuityAuthorizeOptions{}, fmt.Errorf("%s", continuityAuthorizeUsage)
		}
		value := args[index+1]
		switch args[index] {
		case "--decision":
			if options.decision != "" {
				return continuityAuthorizeOptions{}, fmt.Errorf("%s", continuityAuthorizeUsage)
			}
			options.decision = value
		case "--out":
			if options.outputDir != "" {
				return continuityAuthorizeOptions{}, fmt.Errorf("%s", continuityAuthorizeUsage)
			}
			options.outputDir = value
		default:
			return continuityAuthorizeOptions{}, fmt.Errorf("%s", continuityAuthorizeUsage)
		}
		index++
	}
	decision := strings.ToUpper(options.decision)
	if (decision != publiccontinuity.DecisionAccept && decision != publiccontinuity.DecisionReject) || options.outputDir == "" {
		return continuityAuthorizeOptions{}, fmt.Errorf("%s", continuityAuthorizeUsage)
	}
	options.decision = decision
	return options, nil
}

type continuityArtifact struct {
	name string
	data []byte
}

func writeContinuityArtifacts(outputDir string, artifacts []continuityArtifact) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("caller-owned output directory must be empty")
	}
	writes := make([]atomicWrite, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.name == "" || len(artifact.data) == 0 {
			return fmt.Errorf("continuity artifact is empty")
		}
		writes = append(writes, atomicWrite{path: filepath.Join(outputDir, artifact.name), data: artifact.data})
	}
	return writeAtomicFiles(writes)
}

func marshalContinuityJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func continuityWallMS(started time.Time) int64 {
	return int64((time.Since(started).Nanoseconds() + int64(time.Millisecond) - 1) / int64(time.Millisecond))
}
