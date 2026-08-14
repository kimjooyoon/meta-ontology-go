package provenance

import (
	"strings"
	"time"
)

// ContractVersion identifies the executable storage contract.
const ContractVersion = SchemaVersion

// ContractSpec is a compact, machine-readable description of the ledger
// boundary. It intentionally contains no CI or GitHub authority vocabulary.
type ContractSpec struct {
	Version       int
	Format        string
	Required      []string
	Integrity     []string
	Statuses      []string
	NegativeCases []string
	Deferred      []string
}

// CurrentContract returns fresh slices so callers may annotate a copy.
func CurrentContract() ContractSpec {
	return ContractSpec{
		Version: ContractVersion,
		Format:  "one compact UTF-8 JSON object per LF-terminated canonical line",
		Required: []string{
			"schema", "id", "semantic_id", "producer", "kind", "status",
			"source_span", "source_digest", "semantic_digest", "graph_digest",
			"predecessor", "sequence", "freshness", "hash",
		},
		Integrity: []string{
			"sha256 canonical content hash", "contiguous predecessor ID/digest chain",
			"durable two-phase commit metadata with exact base/post bytes",
			"prepared-transaction rollback", "unknown-field rejection", "source freshness",
		},
		Statuses:      []string{"verified", "candidate", "deferred", "failed", "rejected"},
		NegativeCases: []string{"duplicate-conflict", "digest-conflict", "stale-source", "unknown-field", "malformed", "chain-gap", "mutation", "reorder", "truncation", "commit-metadata-missing", "partial-batch-authority", "candidate-as-verified"},
		Deferred:      []string{"GitHub or credential inference", "CI publishing", "business authority inference"},
	}
}

// BillingFixture returns two deterministic source-backed events for the
// repository's billing example. Store.Append supplies their chain links and
// content hashes without consulting any external authority.
func BillingFixture() []Evidence {
	source := strings.Repeat("a", 64)
	semantic := strings.Repeat("b", 64)
	graph := strings.Repeat("c", 64)
	produced := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	return []Evidence{
		{
			ID: "billing://event/order-created", SemanticID: "billing://entity/order",
			Producer: "gooo://producer/billing-fixture", Kind: KindObservation, Status: StatusVerified,
			SourceSpan:   SourceSpan{URI: "examples/billing/main.gooo", Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 41, Line: 2, Column: 1}},
			SourceDigest: source, SemanticDigest: semantic, GraphDigest: graph,
			Freshness:  NewFreshness(source, produced, produced.Add(2*time.Hour)),
			Attributes: map[string]string{"fixture": "billing", "status": "verified"},
		},
		{
			ID: "billing://event/payment-created", SemanticID: "billing://entity/payment",
			Producer: "gooo://producer/billing-fixture", Kind: KindObservation, Status: StatusVerified,
			SourceSpan:   SourceSpan{URI: "examples/billing/main.gooo", Start: Position{Offset: 42, Line: 2, Column: 1}, End: Position{Offset: 88, Line: 3, Column: 1}},
			SourceDigest: source, SemanticDigest: semantic, GraphDigest: graph,
			Freshness:  NewFreshness(source, produced, produced.Add(2*time.Hour)),
			Attributes: map[string]string{"fixture": "billing", "status": "verified"},
		},
	}
}
