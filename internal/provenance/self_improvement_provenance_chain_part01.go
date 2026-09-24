package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	SelfImprovementProvenanceChainBoundPart01          = "BOUND"
	SelfImprovementProvenanceChainUnknownPart01        = "UNKNOWN"
	SelfImprovementProvenanceChainCompleteReasonPart01 = "PROVENANCE_CHAIN_COMPLETE"
	SelfImprovementProvenanceChainNonAuthorizingPart01 = true
)

type SelfImprovementProvenanceStagePart01 struct {
	Name   string
	Digest string
	Bound  bool
}

type SelfImprovementProvenanceChainPart01 struct {
	DeclarationDigest        string
	SourceDigest             string
	SemanticIRDigest         string
	GraphDigest              string
	GeneratedDigest          string
	ReverseObservationDigest string
	Stages                   []SelfImprovementProvenanceStagePart01
	Status                   string
	CausalReason             string
	ChainDigest              string
	AdoptionAuthorized       bool
	NonAuthorizing           bool
}

func BuildSelfImprovementProvenanceChainPart01(
	declarationDigest,
	sourceDigest,
	semanticIRDigest,
	graphDigest,
	generatedDigest,
	reverseObservationDigest string,
) SelfImprovementProvenanceChainPart01 {
	values := []struct {
		name   string
		digest string
	}{
		{"gooo-declaration", normalizeSelfImprovementProvenanceDigestPart01(declarationDigest)},
		{"source", normalizeSelfImprovementProvenanceDigestPart01(sourceDigest)},
		{"semantic-ir", normalizeSelfImprovementProvenanceDigestPart01(semanticIRDigest)},
		{"semantic-graph", normalizeSelfImprovementProvenanceDigestPart01(graphDigest)},
		{"generated", normalizeSelfImprovementProvenanceDigestPart01(generatedDigest)},
		{"reverse-observation", normalizeSelfImprovementProvenanceDigestPart01(reverseObservationDigest)},
	}
	stages := make([]SelfImprovementProvenanceStagePart01, 0, len(values))
	status := SelfImprovementProvenanceChainBoundPart01
	reason := SelfImprovementProvenanceChainCompleteReasonPart01
	for _, value := range values {
		bound := value.digest != ""
		stages = append(stages, SelfImprovementProvenanceStagePart01{
			Name:   value.name,
			Digest: value.digest,
			Bound:  bound,
		})
		if !bound && status == SelfImprovementProvenanceChainBoundPart01 {
			status = SelfImprovementProvenanceChainUnknownPart01
			reason = "MISSING_" + strings.ToUpper(strings.ReplaceAll(value.name, "-", "_")) + "_DIGEST"
		}
	}
	chain := SelfImprovementProvenanceChainPart01{
		DeclarationDigest:        values[0].digest,
		SourceDigest:             values[1].digest,
		SemanticIRDigest:         values[2].digest,
		GraphDigest:              values[3].digest,
		GeneratedDigest:          values[4].digest,
		ReverseObservationDigest: values[5].digest,
		Stages:                   stages,
		Status:                   status,
		CausalReason:             reason,
		AdoptionAuthorized:       false,
		NonAuthorizing:           SelfImprovementProvenanceChainNonAuthorizingPart01,
	}
	chain.ChainDigest = hashSelfImprovementProvenanceChainPart01(chain)
	return chain
}

func (chain SelfImprovementProvenanceChainPart01) Validate() error {
	if !chain.NonAuthorizing {
		return fmt.Errorf("self-improvement provenance chain must remain non-authorizing")
	}
	if chain.AdoptionAuthorized {
		return fmt.Errorf("self-improvement provenance chain cannot authorize adoption")
	}
	if len(chain.Stages) != 6 {
		return fmt.Errorf("self-improvement provenance chain requires six ordered stages, got %d", len(chain.Stages))
	}
	expectedNames := []string{
		"gooo-declaration",
		"source",
		"semantic-ir",
		"semantic-graph",
		"generated",
		"reverse-observation",
	}
	expectedDigests := []string{
		chain.DeclarationDigest,
		chain.SourceDigest,
		chain.SemanticIRDigest,
		chain.GraphDigest,
		chain.GeneratedDigest,
		chain.ReverseObservationDigest,
	}
	for index, stage := range chain.Stages {
		if stage.Name != expectedNames[index] {
			return fmt.Errorf("self-improvement provenance stage %d is %q, want %q", index, stage.Name, expectedNames[index])
		}
		if stage.Digest != expectedDigests[index] {
			return fmt.Errorf("self-improvement provenance stage %q does not match its named digest", stage.Name)
		}
		if stage.Bound != (stage.Digest != "") {
			return fmt.Errorf("self-improvement provenance stage %q has inconsistent bound state", stage.Name)
		}
	}
	if chain.Status != SelfImprovementProvenanceChainBoundPart01 &&
		chain.Status != SelfImprovementProvenanceChainUnknownPart01 {
		return fmt.Errorf("unknown self-improvement provenance chain status %q", chain.Status)
	}
	if chain.Status == SelfImprovementProvenanceChainBoundPart01 &&
		chain.CausalReason != SelfImprovementProvenanceChainCompleteReasonPart01 {
		return fmt.Errorf("bound self-improvement provenance chain has unexpected reason %q", chain.CausalReason)
	}
	if chain.Status == SelfImprovementProvenanceChainUnknownPart01 &&
		!strings.HasPrefix(chain.CausalReason, "MISSING_") {
		return fmt.Errorf("unknown self-improvement provenance chain has unexpected reason %q", chain.CausalReason)
	}
	if chain.ChainDigest == "" || chain.ChainDigest != hashSelfImprovementProvenanceChainPart01(chain) {
		return fmt.Errorf("self-improvement provenance chain digest does not match its ordered evidence")
	}
	return nil
}

func normalizeSelfImprovementProvenanceDigestPart01(value string) string {
	return strings.TrimSpace(value)
}

func hashSelfImprovementProvenanceChainPart01(chain SelfImprovementProvenanceChainPart01) string {
	var canonical strings.Builder
	for _, stage := range chain.Stages {
		canonical.WriteString(stage.Name)
		canonical.WriteByte('=')
		canonical.WriteString(stage.Digest)
		canonical.WriteByte(';')
	}
	canonical.WriteString("status=")
	canonical.WriteString(chain.Status)
	canonical.WriteString(";reason=")
	canonical.WriteString(chain.CausalReason)
	canonical.WriteString(";adoption-authorized=false;non-authorizing=true")
	digest := sha256.Sum256([]byte(canonical.String()))
	return hex.EncodeToString(digest[:])
}
