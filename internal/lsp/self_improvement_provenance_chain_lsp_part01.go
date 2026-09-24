package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

type SelfImprovementProvenanceChainProjectionPart01 struct {
	URI                  string                                            `json:"uri"`
	Version              int                                               `json:"version"`
	Status               string                                            `json:"status"`
	CausalReason         string                                            `json:"causal_reason"`
	ChainDigest          string                                            `json:"chain_digest"`
	Stages               []provenance.SelfImprovementProvenanceStagePart01 `json:"stages"`
	BoundStages          int                                               `json:"bound_stages"`
	TotalStages          int                                               `json:"total_stages"`
	NextRequiredStage    string                                            `json:"next_required_stage"`
	AdoptionAuthorized   bool                                              `json:"adoption_authorized"`
	NonAuthorizing       bool                                              `json:"non_authorizing"`
	EvidencePrefixDigest string                                            `json:"evidence_prefix_digest"`
	MissingStageIndex    int                                               `json:"missing_stage_index"`
	Digest               string                                            `json:"digest"`
}

func ProjectSelfImprovementProvenanceChainPart01(
	uri string,
	version int,
	chain provenance.SelfImprovementProvenanceChainPart01,
) (SelfImprovementProvenanceChainProjectionPart01, error) {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return SelfImprovementProvenanceChainProjectionPart01{}, fmt.Errorf("self-improvement provenance chain projection requires a URI")
	}
	if version < 0 {
		return SelfImprovementProvenanceChainProjectionPart01{}, fmt.Errorf("self-improvement provenance chain projection version cannot be negative")
	}
	if err := chain.Validate(); err != nil {
		return SelfImprovementProvenanceChainProjectionPart01{}, fmt.Errorf("validate self-improvement provenance chain: %w", err)
	}
	boundStages := 0
	for _, stage := range chain.Stages {
		if stage.Bound {
			boundStages++
		}
	}
	projection := SelfImprovementProvenanceChainProjectionPart01{
		URI:                uri,
		Version:            version,
		Status:             chain.Status,
		CausalReason:       chain.CausalReason,
		ChainDigest:        chain.ChainDigest,
		Stages:             append([]provenance.SelfImprovementProvenanceStagePart01(nil), chain.Stages...),
		BoundStages:        boundStages,
		TotalStages:        len(chain.Stages),
		NextRequiredStage:  nextSelfImprovementProvenanceStagePart01(chain),
		AdoptionAuthorized: false,
		NonAuthorizing:     true,
		MissingStageIndex:  missingStageIndexPart01(chain.Stages),
	}
	projection.EvidencePrefixDigest = evidencePrefixDigestPart01(projection)
	projection.Digest = digestSelfImprovementProvenanceChainProjectionPart01(projection)
	return projection, nil
}

func (projection SelfImprovementProvenanceChainProjectionPart01) Validate() error {
	if strings.TrimSpace(projection.URI) == "" {
		return fmt.Errorf("self-improvement provenance chain projection has no URI")
	}
	if projection.Version < 0 {
		return fmt.Errorf("self-improvement provenance chain projection version cannot be negative")
	}
	if projection.AdoptionAuthorized {
		return fmt.Errorf("self-improvement provenance chain projection cannot authorize adoption")
	}
	if !projection.NonAuthorizing {
		return fmt.Errorf("self-improvement provenance chain projection must remain non-authorizing")
	}
	if projection.Status != provenance.SelfImprovementProvenanceChainBoundPart01 &&
		projection.Status != provenance.SelfImprovementProvenanceChainUnknownPart01 {
		return fmt.Errorf("unknown self-improvement provenance chain projection status %q", projection.Status)
	}
	if projection.TotalStages != len(projection.Stages) {
		return fmt.Errorf("self-improvement provenance chain projection total stage count does not match its stages")
	}
	boundStages := 0
	for _, stage := range projection.Stages {
		if stage.Bound {
			boundStages++
		}
	}
	if projection.BoundStages != boundStages {
		return fmt.Errorf("self-improvement provenance chain projection bound stage count does not match its stages")
	}
	if projection.MissingStageIndex != missingStageIndexPart01(projection.Stages) {
		return fmt.Errorf("self-improvement provenance chain projection missing stage index does not match its stages")
	}
	if projection.Status == provenance.SelfImprovementProvenanceChainBoundPart01 &&
		projection.NextRequiredStage != "" {
		return fmt.Errorf("bound self-improvement provenance chain projection has a next required stage")
	}
	if projection.Status == provenance.SelfImprovementProvenanceChainUnknownPart01 &&
		projection.NextRequiredStage == "" {
		return fmt.Errorf("unknown self-improvement provenance chain projection has no next required stage")
	}
	if projection.ChainDigest == "" || projection.EvidencePrefixDigest == "" ||
		projection.EvidencePrefixDigest != evidencePrefixDigestPart01(projection) ||
		projection.Digest == "" ||
		projection.Digest != digestSelfImprovementProvenanceChainProjectionPart01(projection) {
		return fmt.Errorf("self-improvement provenance chain projection digest does not match its evidence")
	}
	return nil
}

func nextSelfImprovementProvenanceStagePart01(
	chain provenance.SelfImprovementProvenanceChainPart01,
) string {
	return nextSelfImprovementProvenanceStagePart01FromStages(chain.Stages)
}

func nextSelfImprovementProvenanceStagePart01FromStages(
	stages []provenance.SelfImprovementProvenanceStagePart01,
) string {
	for _, stage := range stages {
		if !stage.Bound {
			return stage.Name
		}
	}
	return ""
}

func missingStageIndexPart01(
	stages []provenance.SelfImprovementProvenanceStagePart01,
) int {
	for index, stage := range stages {
		if !stage.Bound {
			return index
		}
	}
	return -1
}

func evidencePrefixDigestPart01(
	projection SelfImprovementProvenanceChainProjectionPart01,
) string {
	var canonical strings.Builder
	fmt.Fprintf(
		&canonical,
		"uri=%s;version=%d;status=%s;reason=%s;chain=%s;missing=%d;",
		projection.URI,
		projection.Version,
		projection.Status,
		projection.CausalReason,
		projection.ChainDigest,
		projection.MissingStageIndex,
	)
	for index, stage := range projection.Stages {
		if projection.MissingStageIndex >= 0 && index >= projection.MissingStageIndex {
			break
		}
		fmt.Fprintf(&canonical, "%s=%s;", stage.Name, stage.Digest)
	}
	digest := sha256.Sum256([]byte(canonical.String()))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestSelfImprovementProvenanceChainProjectionPart01(
	projection SelfImprovementProvenanceChainProjectionPart01,
) string {
	projection.Digest = ""
	encoded, _ := json.Marshal(projection)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}
