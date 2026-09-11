package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
)

func main() {
	mode := flag.String("mode", "", "request, resolve, verify, or cases")
	contract := flag.String("contract", "examples/self-improvement/authorization.gooo", "authorization contract")
	candidate := flag.String("candidate", "", "candidate report JSON")
	metadata := flag.String("metadata", "", "candidate artifact metadata JSON")
	request := flag.String("request", "", "authorization request JSON")
	decision := flag.String("decision", "", "explicit decision input JSON")
	resolution := flag.String("resolution", "", "authorization resolution JSON")
	output := flag.String("output", "", "output JSON")
	check := flag.Bool("check", false, "require independent validation")
	flag.Parse()

	if err := run(*mode, *contract, *candidate, *metadata, *request, *decision, *resolution, *output, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(mode, contractPath, candidatePath, metadataPath, requestPath, decisionPath, resolutionPath, outputPath string, check bool) error {
	if outputPath == "" {
		return errors.New("authorization output path is required")
	}
	switch mode {
	case "request":
		return runRequest(contractPath, candidatePath, metadataPath, outputPath, check)
	case "resolve":
		return runResolve(requestPath, decisionPath, outputPath, check)
	case "verify":
		return runVerify(requestPath, resolutionPath, outputPath, check)
	case "cases":
		return runCases(requestPath, outputPath, check)
	default:
		return errors.New("authorization mode must be request, resolve, verify, or cases")
	}
}

func runRequest(contractPath, candidatePath, metadataPath, outputPath string, check bool) error {
	metadata, err := readMetadata(metadataPath)
	if err != nil {
		return err
	}
	candidateRaw, err := os.ReadFile(candidatePath)
	if err != nil {
		return writeJSON(outputPath, selfimprovementcandidate.BuildUnavailableAuthorizationResolution(metadata))
	}
	request, err := selfimprovementcandidate.BuildAuthorizationRequest(os.DirFS("."), contractPath, candidateRaw, metadata)
	if err != nil {
		return fmt.Errorf("build authorization request: %w", err)
	}
	if check {
		if err := selfimprovementcandidate.ValidateAuthorizationRequest(request); err != nil {
			return err
		}
	}
	return writeJSON(outputPath, request)
}

func runResolve(requestPath, decisionPath, outputPath string, check bool) error {
	request, err := readRequest(requestPath)
	if err != nil {
		return err
	}
	var inputs []selfimprovementcandidate.AuthorizationDecisionInput
	if decisionPath != "" {
		input, err := readDecision(decisionPath)
		if err != nil {
			return err
		}
		inputs = []selfimprovementcandidate.AuthorizationDecisionInput{input}
	}
	resolution := selfimprovementcandidate.ResolveAuthorization(request, inputs)
	if check {
		if err := selfimprovementcandidate.VerifyAuthorizationResolution(request, resolution); err != nil {
			return err
		}
	}
	return writeJSON(outputPath, resolution)
}

func runVerify(requestPath, resolutionPath, outputPath string, check bool) error {
	request, err := readRequest(requestPath)
	if err != nil {
		return err
	}
	resolution, err := readResolution(resolutionPath)
	if err != nil {
		return err
	}
	verification, err := selfimprovementcandidate.BuildAuthorizationVerification(request, resolution)
	if err != nil {
		return err
	}
	if check && !verification.DecisionVerified {
		return errors.New("authorization decision was not independently verified")
	}
	return writeJSON(outputPath, verification)
}

func runCases(requestPath, outputPath string, check bool) error {
	request, err := readRequest(requestPath)
	if err != nil {
		return err
	}
	report, err := selfimprovementcandidate.BuildCanonicalCases(request)
	if err != nil {
		return err
	}
	if check {
		if err := selfimprovementcandidate.ValidateCanonicalCases(report); err != nil {
			return err
		}
	}
	return writeJSON(outputPath, report)
}

func readMetadata(path string) (selfimprovementcandidate.ArtifactMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return selfimprovementcandidate.ArtifactMetadata{}, fmt.Errorf("read artifact metadata: %w", err)
	}
	var metadata selfimprovementcandidate.ArtifactMetadata
	if err := decode(data, &metadata); err != nil {
		return metadata, fmt.Errorf("decode artifact metadata: %w", err)
	}
	return metadata, nil
}

func readRequest(path string) (selfimprovementcandidate.AuthorizationRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return selfimprovementcandidate.AuthorizationRequest{}, fmt.Errorf("read authorization request: %w", err)
	}
	var request selfimprovementcandidate.AuthorizationRequest
	if err := decode(data, &request); err != nil {
		return request, fmt.Errorf("decode authorization request: %w", err)
	}
	if err := selfimprovementcandidate.ValidateAuthorizationRequest(request); err != nil {
		return request, err
	}
	return request, nil
}

func readDecision(path string) (selfimprovementcandidate.AuthorizationDecisionInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return selfimprovementcandidate.AuthorizationDecisionInput{}, fmt.Errorf("read decision input: %w", err)
	}
	var input selfimprovementcandidate.AuthorizationDecisionInput
	if err := decode(data, &input); err != nil {
		return input, fmt.Errorf("decode decision input: %w", err)
	}
	input.DecisionDigest = decisionInputDigest(data)
	return input, nil
}

func readResolution(path string) (selfimprovementcandidate.AuthorizationResolution, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return selfimprovementcandidate.AuthorizationResolution{}, fmt.Errorf("read authorization resolution: %w", err)
	}
	var resolution selfimprovementcandidate.AuthorizationResolution
	if err := decode(data, &resolution); err != nil {
		return resolution, fmt.Errorf("decode authorization resolution: %w", err)
	}
	return resolution, nil
}

func decisionInputDigest(data []byte) string {
	// The package's canonical decision digest is deliberately reconstructed by
	// the evaluator; this raw digest is only provenance metadata for the CLI.
	var input selfimprovementcandidate.AuthorizationDecisionInput
	if err := decode(data, &input); err != nil {
		return ""
	}
	input.DecisionDigest = ""
	canonical, _ := json.Marshal(input)
	return digestBytes(canonical)
}

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func decode(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("JSON contains trailing content")
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode authorization artifact: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write authorization artifact: %w", err)
	}
	return nil
}
