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
	URI                string                                      `json:"uri"`
	Version            int                                         `json:"version"`
	Status             string                                      `json:"status"`
	CausalReason       string                                      `json:"causal_reason"`
	ChainDigest        string                                      `json:"chain_digest"`
	Stages             []provenance.SelfImprovementProvenanceStagePart01 `json:"stages"`
	AdoptionAuthorized bool                                        `json:"adoption_authorized"`
	NonAuthorizing     bool                                        `json:"non_authorizing"`
	Digest             string                                      `json:"digest"`
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
	projection := SelfImprovementProvenanceChainProjectionPart01{
		URI:                uri,
		Version:            version,
		Status:             chain.Status,
		CausalReason:       chain.CausalReason,
		ChainDigest:        chain.ChainDigest,
		Stages:             append([]provenance.SelfImprovementProvenanceStagePart01(nil), chain.Stages...),
		AdoptionAuthorized: false,
		NonAuthorizing:     true,
	}
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
	if projection.ChainDigest == "" || projection.Digest == "" ||
		projection.Digest != digestSelfImprovementProvenanceChainProjectionPart01(projection) {
		return fmt.Errorf("self-improvement provenance chain projection digest does not match its evidence")
	}
	return nil
}

func digestSelfImprovementProvenanceChainProjectionPart01(
	projection SelfImprovementProvenanceChainProjectionPart01,
) string {
	projection.Digest = ""
	encoded, _ := json.Marshal(projection)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}
