package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type inputSpec struct {
	id   string
	path string
}

func loadArtifacts(opts options) ([]artifactEvidence, provenanceEnvelope, sourceInventory, error) {
	specs := []inputSpec{
		{id: "source-metrics", path: opts.MetricsPath},
		{id: "self-improvement-plan", path: opts.PlanPath},
		{id: "self-improvement-execution", path: opts.ExecutionPath},
		{id: "self-improvement-receipts", path: opts.ReceiptsPath},
		{id: "artifact-provenance", path: opts.ProvenancePath},
	}
	artifacts := make([]artifactEvidence, 0, len(specs))
	var provenance provenanceEnvelope
	var inventory sourceInventory
	for _, spec := range specs {
		data, err := os.ReadFile(spec.path)
		if err != nil {
			return nil, provenance, inventory, fmt.Errorf("read %s: %w", spec.id, err)
		}
		if len(data) == 0 {
			return nil, provenance, inventory, fmt.Errorf("%s is empty", spec.id)
		}
		sum := sha256.Sum256(data)
		artifacts = append(artifacts, artifactEvidence{
			ID: spec.id, File: filepath.Base(spec.path), Bytes: len(data), SHA256: hex.EncodeToString(sum[:]),
		})
		if spec.id == "artifact-provenance" {
			provenance, err = decodeProvenance(data)
			if err != nil {
				return nil, provenance, inventory, err
			}
		}
		if spec.id == "source-metrics" {
			inventory, err = decodeSourceInventory(data)
			if err != nil {
				return nil, provenance, inventory, err
			}
		}
		if spec.id == "self-improvement-plan" {
			inventory.SelectedSubjects, err = decodeSelectedSubjects(data)
			if err != nil {
				return nil, provenance, inventory, err
			}
			inventory.SelectedOperations = len(inventory.SelectedSubjects)
		}
	}
	return artifacts, provenance, inventory, nil
}

func digestArtifacts(artifacts []artifactEvidence) string {
	hash := sha256.New()
	for _, artifact := range artifacts {
		_, _ = io.WriteString(hash, artifact.ID)
		_, _ = hash.Write([]byte{0})
		_, _ = io.WriteString(hash, artifact.SHA256)
		_, _ = hash.Write([]byte{'\n'})
	}
	return hex.EncodeToString(hash.Sum(nil))
}
