package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"strings"
)

// OriginChainStage identifies one measured boundary in a language-to-runtime
// provenance chain. It is descriptive evidence, not an authorization grant.
type OriginChainStage string

const (
	OriginChainStageDeclaration        OriginChainStage = "DECLARATION"
	OriginChainStageIR                 OriginChainStage = "IR"
	OriginChainStageGeneration         OriginChainStage = "GENERATION"
	OriginChainStageReverseObservation OriginChainStage = "REVERSE_OBSERVATION"
	OriginChainStageMetric             OriginChainStage = "METRIC"
)

// OriginChainStatus describes whether all required origin links are present.
type OriginChainStatus string

const (
	OriginChainStatusUnknown  OriginChainStatus = "UNKNOWN"
	OriginChainStatusPartial  OriginChainStatus = "PARTIAL"
	OriginChainStatusComplete OriginChainStatus = "COMPLETE"
)

// OriginChain carries source-backed links from a .gooo declaration through
// the executable representation and back to a measured observation.
// EvidenceDigest is supplied by the producer of the upstream observation; it
// is never inferred from the other fields.
type OriginChain struct {
	DeclarationURI        string `json:"declaration_uri"`
	DeclarationSymbol     string `json:"declaration_symbol"`
	IRNode                string `json:"ir_node"`
	GeneratedURI          string `json:"generated_uri"`
	GeneratedSymbol       string `json:"generated_symbol"`
	ReverseObservationURI string `json:"reverse_observation_uri"`
	ReverseObservation    string `json:"reverse_observation"`
	MetricName            string `json:"metric_name"`
	MetricValue           string `json:"metric_value"`
	EvidenceDigest        string `json:"evidence_digest"`
}

// OriginChainObservation is a value-level result of inspecting an origin
// chain. A PARTIAL or UNKNOWN result is deliberately not comparable as proof.
type OriginChainObservation struct {
	Chain   OriginChain        `json:"chain"`
	Status  OriginChainStatus  `json:"status"`
	Missing []OriginChainStage `json:"missing,omitempty"`
	Reason  string             `json:"reason,omitempty"`
	Digest  string             `json:"digest,omitempty"`
}

// OriginChainTransition reports only an evidence-backed digest transition.
type OriginChainTransition string

const (
	OriginChainTransitionUnknown   OriginChainTransition = "UNKNOWN"
	OriginChainTransitionUnchanged OriginChainTransition = "UNCHANGED"
	OriginChainTransitionChanged   OriginChainTransition = "CHANGED"
)

// ObserveOriginChain normalizes boundary labels and computes a stable digest.
// It does not parse, execute, authenticate, or infer any missing evidence.
func ObserveOriginChain(chain OriginChain) OriginChainObservation {
	chain = normalizeOriginChain(chain)
	missing := originChainMissing(chain)
	status := OriginChainStatusComplete
	reason := ""
	switch {
	case len(missing) == originChainFieldCount:
		status = OriginChainStatusUnknown
		reason = "origin chain has no linked evidence"
	case len(missing) > 0:
		status = OriginChainStatusPartial
		reason = "origin chain is incomplete"
	}
	return OriginChainObservation{
		Chain: chain, Status: status, Missing: missing, Reason: reason,
		Digest: originChainDigest(chain),
	}
}

// Comparable reports whether the observation has enough measured boundaries
// to support a transition comparison without treating UNKNOWN as unchanged.
func (observation OriginChainObservation) Comparable() bool {
	return observation.Status == OriginChainStatusComplete && observation.Digest != ""
}

// CompareOriginChains compares only complete observations. A changed digest
// is reported as CHANGED, without claiming that the change is beneficial.
func CompareOriginChains(before, after OriginChainObservation) OriginChainTransition {
	if !before.Comparable() || !after.Comparable() {
		return OriginChainTransitionUnknown
	}
	if before.Digest == after.Digest {
		return OriginChainTransitionUnchanged
	}
	return OriginChainTransitionChanged
}

const originChainFieldCount = 5

func normalizeOriginChain(chain OriginChain) OriginChain {
	chain.DeclarationURI = strings.TrimSpace(chain.DeclarationURI)
	chain.DeclarationSymbol = strings.TrimSpace(chain.DeclarationSymbol)
	chain.IRNode = strings.TrimSpace(chain.IRNode)
	chain.GeneratedURI = strings.TrimSpace(chain.GeneratedURI)
	chain.GeneratedSymbol = strings.TrimSpace(chain.GeneratedSymbol)
	chain.ReverseObservationURI = strings.TrimSpace(chain.ReverseObservationURI)
	chain.ReverseObservation = strings.TrimSpace(chain.ReverseObservation)
	chain.MetricName = strings.TrimSpace(chain.MetricName)
	chain.MetricValue = strings.TrimSpace(chain.MetricValue)
	chain.EvidenceDigest = strings.TrimSpace(chain.EvidenceDigest)
	return chain
}

func originChainMissing(chain OriginChain) []OriginChainStage {
	missing := make([]OriginChainStage, 0, originChainFieldCount)
	if chain.DeclarationURI == "" || chain.DeclarationSymbol == "" {
		missing = append(missing, OriginChainStageDeclaration)
	}
	if chain.IRNode == "" {
		missing = append(missing, OriginChainStageIR)
	}
	if chain.GeneratedURI == "" || chain.GeneratedSymbol == "" {
		missing = append(missing, OriginChainStageGeneration)
	}
	if chain.ReverseObservationURI == "" || chain.ReverseObservation == "" {
		missing = append(missing, OriginChainStageReverseObservation)
	}
	if chain.MetricName == "" || chain.MetricValue == "" || chain.EvidenceDigest == "" {
		missing = append(missing, OriginChainStageMetric)
	}
	return missing
}

func originChainDigest(chain OriginChain) string {
	h := sha256.New()
	writeOriginField(h, "declaration_uri", chain.DeclarationURI)
	writeOriginField(h, "declaration_symbol", chain.DeclarationSymbol)
	writeOriginField(h, "ir_node", chain.IRNode)
	writeOriginField(h, "generated_uri", chain.GeneratedURI)
	writeOriginField(h, "generated_symbol", chain.GeneratedSymbol)
	writeOriginField(h, "reverse_observation_uri", chain.ReverseObservationURI)
	writeOriginField(h, "reverse_observation", chain.ReverseObservation)
	writeOriginField(h, "metric_name", chain.MetricName)
	writeOriginField(h, "metric_value", chain.MetricValue)
	writeOriginField(h, "evidence_digest", chain.EvidenceDigest)
	return hex.EncodeToString(h.Sum(nil))
}

func writeOriginField(h hash.Hash, name, value string) {
	h.Write([]byte(name))
	h.Write([]byte{0})
	h.Write([]byte(value))
	h.Write([]byte{0})
}
