package main

import (
	"encoding/json"
	"fmt"
	"os"

	metalsp "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainlsp"
)

type conceptItem struct {
	ID             string            `json:"id"`
	MetaOperation  string            `json:"meta_operation"`
	Stage          string            `json:"stage"`
	CodeBindings   []string          `json:"code_bindings"`
	MetricBindings []string          `json:"metric_bindings"`
	UseCases       []json.RawMessage `json:"use_cases"`
}

type conceptArtifact struct {
	Decision       string `json:"decision"`
	ArtifactDigest string `json:"artifact_digest"`
	Report         struct {
		Concepts []conceptItem `json:"concepts"`
	} `json:"report"`
}

func readConcept(path string) (metalsp.ConceptBinding, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return metalsp.ConceptBinding{}, err
	}
	var artifact conceptArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return metalsp.ConceptBinding{}, err
	}
	var found *conceptItem
	for index := range artifact.Report.Concepts {
		item := &artifact.Report.Concepts[index]
		if item.ID == "toolchain-lsp" {
			if found != nil {
				return metalsp.ConceptBinding{}, fmt.Errorf("duplicate toolchain-lsp concept")
			}
			found = item
		}
	}
	if found == nil {
		return metalsp.ConceptBinding{}, fmt.Errorf("toolchain-lsp concept is missing")
	}
	return metalsp.ConceptBinding{ArtifactDecision: artifact.Decision, ArtifactDigest: artifact.ArtifactDigest,
		ConceptID: found.ID, MetaOperation: found.MetaOperation, Stage: found.Stage,
		CodeBindings: found.CodeBindings, MetricBindings: found.MetricBindings, UseCaseBindings: len(found.UseCases)}, nil
}
