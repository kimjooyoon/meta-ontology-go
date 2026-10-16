package main

import (
	"encoding/json"
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

type artifactMetadata struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Digest      string `json:"digest"`
	Expired     bool   `json:"expired"`
	WorkflowRun struct {
		ID      int64  `json:"id"`
		HeadSHA string `json:"head_sha"`
	} `json:"workflow_run"`
}

func boundArchive(cfg config) ([]byte, error) {
	o := cfg.Origin
	digest, err := canonicalDigest(o.ArtifactDigest)
	if err != nil {
		return nil, err
	}
	if !artifact.ValidSubject(o.Repository, o.SubjectSHA) || o.RunID < 1 || o.Attempt != 1 || o.ArtifactID < 1 {
		return nil, fmt.Errorf("counterexamples require an exact original source/run/attempt/artifact")
	}
	metadataRaw, err := readBounded(cfg.Metadata, 256<<10)
	if err != nil {
		return nil, err
	}
	var metadata artifactMetadata
	if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
		return nil, err
	}
	name := fmt.Sprintf("metric-concept-operation-%s-%d-%d", o.SubjectSHA, o.RunID, o.Attempt)
	metadataDigest, digestErr := canonicalDigest(metadata.Digest)
	if metadata.ID != o.ArtifactID || metadata.Name != name || digestErr != nil || metadataDigest != digest ||
		metadata.Expired || metadata.WorkflowRun.ID != o.RunID || metadata.WorkflowRun.HeadSHA != o.SubjectSHA {
		return nil, fmt.Errorf("producer artifact metadata differs from caller-owned identity")
	}
	raw, err := readBounded(cfg.Archive, 32<<20)
	if err != nil {
		return nil, err
	}
	if artifact.ContentDigest(raw) != digest {
		return nil, fmt.Errorf("producer archive differs from job-owned digest")
	}
	return raw, nil
}
