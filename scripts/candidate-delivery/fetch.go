package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/candidatedelivery"
)

func fetch(ctx context.Context, o options) error {
	input := candidatedelivery.Selection{Repository: o.repository, RunID: o.run, Attempt: o.attempt}
	receipt := candidatedelivery.Unavailable(input, "DELIVERY_API_LOOKUP_UNKNOWN")
	var bundle candidatedelivery.Bundle
	if err := os.MkdirAll(filepath.Dir(o.output), 0700); err != nil {
		return err
	}
	// Each fetch owns a fresh directory, so UNKNOWN cannot expose an older payload.
	if err := os.Mkdir(o.output, 0700); err != nil {
		return err
	}
	run, artifacts, err := lookup(ctx, o)
	if err == nil {
		receipt = candidatedelivery.Select(run, artifacts, input)
		if receipt.State == "LOCATED" {
			archive, downloadErr := get(ctx, fmt.Sprintf("repos/%s/actions/artifacts/%d/zip", o.repository, receipt.Artifact.ID))
			if downloadErr != nil {
				archive = nil
			}
			receipt, bundle = candidatedelivery.Consume(receipt, archive)
		}
	}
	if err := writeJSON(filepath.Join(o.output, "delivery.json"), receipt); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(o.output, "metadata.json"), candidatedelivery.Metadata(receipt)); err != nil {
		return err
	}
	if receipt.State == "CLOSED" {
		for name, raw := range map[string][]byte{"candidate.json": bundle.Candidate,
			"transport.json": bundle.Transport, "manifest.json": bundle.Manifest} {
			if err := os.WriteFile(filepath.Join(o.output, name), raw, 0600); err != nil {
				return err
			}
		}
	}
	if receipt.State == "REFUTED" {
		return fmt.Errorf("%s", receipt.Reason)
	}
	return nil
}

func lookup(ctx context.Context, o options) ([]byte, []byte, error) {
	if o.run <= 0 || o.attempt <= 0 || !repositoryName(o.repository) {
		return nil, nil, fmt.Errorf("invalid delivery input")
	}
	raw, err := get(ctx, fmt.Sprintf("repos/%s/actions/runs/%d", o.repository, o.run))
	if err != nil {
		return nil, nil, err
	}
	var all struct {
		Total     int                          `json:"total_count"`
		Artifacts []candidatedelivery.Artifact `json:"artifacts"`
	}
	all.Artifacts = []candidatedelivery.Artifact{}
	for page := 1; page <= 8; page++ {
		response, err := get(ctx, fmt.Sprintf("repos/%s/actions/runs/%d/artifacts?per_page=100&page=%d", o.repository, o.run, page))
		if err != nil {
			return nil, nil, err
		}
		var part struct {
			Total     int                          `json:"total_count"`
			Artifacts []candidatedelivery.Artifact `json:"artifacts"`
		}
		if err := json.Unmarshal(response, &part); err != nil {
			return nil, nil, err
		}
		if page > 1 && all.Total != part.Total {
			return nil, nil, fmt.Errorf("delivery artifact population changed")
		}
		all.Total = part.Total
		all.Artifacts = append(all.Artifacts, part.Artifacts...)
		if len(all.Artifacts) >= all.Total {
			break
		}
	}
	artifacts, err := json.Marshal(all)
	return raw, artifacts, err
}
