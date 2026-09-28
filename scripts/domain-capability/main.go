package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type observation struct {
	Domain          string `json:"domain"`
	CapabilityID    string `json:"capability_id"`
	Outcome         string `json:"outcome"`
	EvidenceDigest  string `json:"evidence_digest"`
	NonExecuting    bool   `json:"non_executing"`
	NonAuthorizing  bool   `json:"non_authorizing"`
}

type inputDocument struct {
	Version              string        `json:"version"`
	Domain               string        `json:"domain"`
	ExpectedCapabilities []string      `json:"expected_capabilities"`
	Observations         []observation `json:"observations"`
}

type capabilitySummary struct {
	CapabilityID   string   `json:"capability_id"`
	ObservationCount int     `json:"observation_count"`
	UsefulCount      int     `json:"useful_count"`
	NotUsefulCount  int     `json:"not_useful_count"`
	UnresolvedCount int     `json:"unresolved_count"`
	EvidenceDigests []string `json:"evidence_digests"`
}

type report struct {
	Version              string             `json:"version"`
	Domain               string             `json:"domain"`
	ExpectedCapabilities []string           `json:"expected_capabilities"`
	ObservedCapabilities []string           `json:"observed_capabilities"`
	MissingCapabilities  []string           `json:"missing_capabilities"`
	CoverageRatio        float64            `json:"coverage_ratio"`
	UsefulRatio          float64            `json:"useful_ratio"`
	UnresolvedRatio      float64            `json:"unresolved_ratio"`
	ObservationCount     int                `json:"observation_count"`
	Summaries            []capabilitySummary `json:"summaries"`
	DecisionSignal       string             `json:"decision_signal"`
	MeasurementSemantics string             `json:"measurement_semantics"`
	EvidenceDigests      []string           `json:"evidence_digests"`
}

func decodeInput(r io.Reader) (inputDocument, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var input inputDocument
	if err := decoder.Decode(&input); err != nil {
		return inputDocument{}, fmt.Errorf("decode input: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return inputDocument{}, errors.New("trailing JSON is not allowed")
		}
		return inputDocument{}, fmt.Errorf("decode trailing JSON: %w", err)
	}
	return input, nil
}

func validateInput(input inputDocument) error {
	if strings.TrimSpace(input.Version) == "" {
		return errors.New("input version is required")
	}
	if strings.TrimSpace(input.Domain) == "" {
		return errors.New("domain is required")
	}
	if len(input.ExpectedCapabilities) == 0 {
		return errors.New("at least one expected capability is required")
	}

	seen := make(map[string]struct{}, len(input.ExpectedCapabilities))
	for _, capability := range input.ExpectedCapabilities {
		capability = strings.TrimSpace(capability)
		if capability == "" {
			return errors.New("expected capabilities cannot be empty")
		}
		if _, ok := seen[capability]; ok {
			return fmt.Errorf("duplicate expected capability %q", capability)
		}
		seen[capability] = struct{}{}
	}

	expected := seen
	for index, item := range input.Observations {
		if strings.TrimSpace(item.Domain) != input.Domain {
			return fmt.Errorf("observation %d has domain %q, want %q", index, item.Domain, input.Domain)
		}
		if _, ok := expected[item.CapabilityID]; !ok {
			return fmt.Errorf("observation %d references capability %q outside the domain contract", index, item.CapabilityID)
		}
		if item.EvidenceDigest == "" {
			return fmt.Errorf("observation %d is missing evidence digest", index)
		}
		switch item.Outcome {
		case "useful", "not_useful", "unresolved":
		default:
			return fmt.Errorf("observation %d has unsupported outcome %q", index, item.Outcome)
		}
		if !item.NonExecuting || !item.NonAuthorizing {
			return fmt.Errorf("observation %d crossed an execution or authorization boundary", index)
		}
	}
	return nil
}

func buildReport(input inputDocument) (report, error) {
	if err := validateInput(input); err != nil {
		return report{}, err
	}

	expected := append([]string(nil), input.ExpectedCapabilities...)
	sort.Strings(expected)

	grouped := make(map[string]*capabilitySummary, len(expected))
	for _, capability := range expected {
		grouped[capability] = &capabilitySummary{CapabilityID: capability}
	}
	evidenceSeen := make(map[string]struct{})
	observed := make(map[string]struct{})

	var usefulCount, unresolvedCount int
	for _, item := range input.Observations {
		summary := grouped[item.CapabilityID]
		summary.ObservationCount++
		switch item.Outcome {
		case "useful":
			summary.UsefulCount++
			usefulCount++
		case "not_useful":
			summary.NotUsefulCount++
		case "unresolved":
			summary.UnresolvedCount++
			unresolvedCount++
		}
		observed[item.CapabilityID] = struct{}{}
		evidenceSeen[item.EvidenceDigest] = struct{}{}
	}

	observedCapabilities := make([]string, 0, len(observed))
	for capability := range observed {
		observedCapabilities = append(observedCapabilities, capability)
	}
	sort.Strings(observedCapabilities)

	missingCapabilities := make([]string, 0, len(expected)-len(observed))
	for _, capability := range expected {
		if _, ok := observed[capability]; !ok {
			missingCapabilities = append(missingCapabilities, capability)
		}
	}

	summaries := make([]capabilitySummary, 0, len(expected))
	for _, capability := range expected {
		item := *grouped[capability]
		for digest := range evidenceSeen {
			for _, observation := range input.Observations {
				if observation.CapabilityID == capability && observation.EvidenceDigest == digest {
					item.EvidenceDigests = append(item.EvidenceDigests, digest)
					break
				}
			}
		}
		sort.Strings(item.EvidenceDigests)
		summaries = append(summaries, item)
	}

	observations := len(input.Observations)
	coverageRatio := float64(len(observedCapabilities)) / float64(len(expected))
	var usefulRatio, unresolvedRatio float64
	if observations > 0 {
		usefulRatio = float64(usefulCount) / float64(observations)
		unresolvedRatio = float64(unresolvedCount) / float64(observations)
	}

	decision := "NO_ACTIONABLE_SIGNAL"
	switch {
	case unresolvedCount > 0:
		decision = "COLLECT_MORE_OBSERVATIONS"
	case len(missingCapabilities) > 0:
		decision = "REVIEW_MISSING_CAPABILITIES"
	case usefulCount > 0:
		decision = "REVIEW_INVESTMENT"
	}

	evidenceDigests := make([]string, 0, len(evidenceSeen))
	for digest := range evidenceSeen {
		evidenceDigests = append(evidenceDigests, digest)
	}
	sort.Strings(evidenceDigests)

	return report{
		Version:                "domain.capability.measurement.v1",
		Domain:                 input.Domain,
		ExpectedCapabilities:   expected,
		ObservedCapabilities:   observedCapabilities,
		MissingCapabilities:    missingCapabilities,
		CoverageRatio:          coverageRatio,
		UsefulRatio:            usefulRatio,
		UnresolvedRatio:        unresolvedRatio,
		ObservationCount:       observations,
		Summaries:              summaries,
		DecisionSignal:         decision,
		MeasurementSemantics:   "observational decision signal; not semantic completeness, correctness, or authorization",
		EvidenceDigests:        evidenceDigests,
	}, nil
}

func main() {
	inputPath := flag.String("input", "-", "domain capability JSON path, or - for stdin")
	flag.Parse()

	var input io.Reader = os.Stdin
	var inputFile *os.File
	if *inputPath != "-" {
		var err error
		inputFile, err = os.Open(*inputPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer inputFile.Close()
		input = inputFile
	}

	document, err := decodeInput(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	output, err := buildReport(document)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
