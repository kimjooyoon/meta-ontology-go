package provenance

import (
	"strings"
	"time"
)

func testRecord(id, semanticID string, status EvidenceStatus) Evidence {
	source := strings.Repeat("d", 64)
	return Evidence{
		ID: id, SemanticID: semanticID, Producer: "test://producer/store", Kind: KindVerification, Status: status,
		SourceSpan:   SourceSpan{URI: "examples/billing/main.gooo", Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 10, Line: 1, Column: 11}},
		SourceDigest: source, SemanticDigest: strings.Repeat("e", 64), GraphDigest: strings.Repeat("f", 64),
		Freshness:  NewFreshness(source, time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 14, 2, 0, 0, 0, time.UTC)),
		Attributes: map[string]string{"fixture": "billing", "status": string(status)},
	}
}
