package main

import (
	"encoding/json"
	"os"

	release "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainrelease"
)

type conceptItem struct {
	ID             string   `json:"id"`
	MetaOperation  string   `json:"meta_operation"`
	NoveltyClaim   bool     `json:"novelty_claim"`
	CodeBindings   []string `json:"code_bindings"`
	MetricBindings []string `json:"metric_bindings"`
	UseCases       []any    `json:"use_cases"`
}

type conceptArtifact struct {
	ArtifactDigest string `json:"artifact_digest"`
	Report         struct {
		Concepts []conceptItem `json:"concepts"`
	} `json:"report"`
}

func readConcept(path string) (string, bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", false, err
	}
	var artifact conceptArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return "", false, err
	}
	matches := 0
	for _, concept := range artifact.Report.Concepts {
		if concept.ID != "toolchain-cross-platform-release" {
			continue
		}
		matches++
		if concept.MetaOperation != release.MetaOperation || concept.NoveltyClaim ||
			len(concept.CodeBindings) != 6 ||
			len(concept.MetricBindings) != release.IndicatorCount ||
			len(concept.UseCases) != 3 {
			return artifact.ArtifactDigest, false, nil
		}
	}
	return artifact.ArtifactDigest, matches == 1 && artifact.ArtifactDigest != "", nil
}
