package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const adoptionProposalUsage = "usage: gooo propose <compiler-observation.json> --out <directory>"

type adoptionProposalOptions struct {
	observationFilename string
	outputDir           string
}

func runAdoptionProposal(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseAdoptionProposalArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	data, err := reader.ReadFile(options.observationFilename)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: adoption proposal: read observation: %v\n", err)
		return exitFailure
	}
	var observation generation.SemanticObservation
	if err := json.Unmarshal(data, &observation); err != nil {
		fmt.Fprintf(stderr, "gooo: adoption proposal: decode observation: %v\n", err)
		return exitFailure
	}
	if err := generation.VerifySemanticObservation(observation); err != nil {
		fmt.Fprintf(stderr, "gooo: adoption proposal: observation verification: %v\n", err)
		return exitFailure
	}
	if observation.Decision != "CLOSED" || len(observation.Candidates) != 1 || observation.Candidates[0].ExpectedReducibleCount <= 0 {
		fmt.Fprintln(stderr, "gooo: adoption proposal: observation does not contain exactly one reducible candidate")
		return exitFailure
	}
	proposal := generation.SemanticAdoptionProposal{
		Schema:            generation.SemanticAdoptionProposalSchema,
		ObservationDigest: cache.HashBytes(data).String(),
		ContractDigest:    observation.ContractDigest,
		InputSourceDigest: observation.InputSourceDigest,
		Candidate:         observation.Candidates[0],
		Target:            generation.SemanticAdoptionTarget,
		Mode:              generation.SemanticAdoptionMode,
		ExecutionAllowed:  false,
		RepositoryWrites:  0,
	}
	if err := generation.ValidateSemanticAdoptionProposal(proposal); err != nil {
		fmt.Fprintf(stderr, "gooo: adoption proposal: %v\n", err)
		return exitFailure
	}
	proposalData, err := json.MarshalIndent(proposal, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: adoption proposal: encode: %v\n", err)
		return exitFailure
	}
	proposalData = append(proposalData, '\n')
	if err := writeAdoptionArtifact(options.outputDir, "adoption-proposal.json", proposalData); err != nil {
		fmt.Fprintf(stderr, "gooo: adoption proposal: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "proposal: %s\n", filepath.Join(options.outputDir, "adoption-proposal.json"))
	return exitOK
}

func parseAdoptionProposalArguments(args []string) (adoptionProposalOptions, error) {
	if len(args) == 0 {
		return adoptionProposalOptions{}, fmt.Errorf("%s", adoptionProposalUsage)
	}
	options := adoptionProposalOptions{observationFilename: args[0]}
	for index := 1; index < len(args); index++ {
		if index+1 >= len(args) || args[index+1] == "" {
			return adoptionProposalOptions{}, fmt.Errorf("%s", adoptionProposalUsage)
		}
		switch args[index] {
		case "--out":
			if options.outputDir != "" {
				return adoptionProposalOptions{}, fmt.Errorf("%s", adoptionProposalUsage)
			}
			options.outputDir = args[index+1]
		default:
			return adoptionProposalOptions{}, fmt.Errorf("%s", adoptionProposalUsage)
		}
		index++
	}
	if options.observationFilename == "" || options.outputDir == "" {
		return adoptionProposalOptions{}, fmt.Errorf("%s", adoptionProposalUsage)
	}
	return options, nil
}

func writeAdoptionArtifact(outputDir, filename string, data []byte) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create caller-owned output directory: %w", err)
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("inspect caller-owned output directory: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("caller-owned output directory must be empty")
	}
	if err := os.WriteFile(filepath.Join(outputDir, filename), data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}
	return nil
}
