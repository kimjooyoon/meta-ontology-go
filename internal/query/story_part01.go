package query

import (
	"encoding/json"
	"errors"
	"maps"
	"sort"

	provenance "github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

const StorySchema = "gooo-query-story/v1"

const (
	StoryAvailable     = "available"
	StoryCandidate     = "candidate"
	StoryRejected      = "rejected"
	StoryContradictory = "contradictory"
	StoryUnknown       = "unknown"
)

type StoryEvidenceReference struct {
	ID             string                    `json:"id"`
	SemanticID     string                    `json:"semantic_id"`
	Producer       string                    `json:"producer"`
	Kind           provenance.EvidenceKind   `json:"kind"`
	Status         provenance.EvidenceStatus `json:"status"`
	SourceSpan     provenance.SourceSpan     `json:"source_span"`
	SourceDigest   string                    `json:"source_digest"`
	SemanticDigest string                    `json:"semantic_digest"`
	GraphDigest    string                    `json:"graph_digest"`
	Sequence       uint64                    `json:"sequence"`
	Predecessor    *provenance.DigestLink    `json:"predecessor,omitempty"`
	Attributes     map[string]string         `json:"attributes,omitempty"`
	Freshness      provenance.Freshness      `json:"freshness"`
	Hash           string                    `json:"hash"`
}

type StoryResponse struct {
	Schema         string                   `json:"schema"`
	SemanticID     ID                       `json:"semantic_id"`
	SemanticDigest string                   `json:"semantic_digest,omitempty"`
	GraphHash      string                   `json:"graph_hash"`
	LedgerDigest   string                   `json:"ledger_digest,omitempty"`
	Node           Node                     `json:"node"`
	Relationships  []Fact                   `json:"relationships"`
	Evidence       []StoryEvidenceReference `json:"evidence"`
	Status         string                   `json:"status"`
	Reason         string                   `json:"reason"`
	BindingDigest  string                   `json:"binding_digest"`
}

func (graph Graph) Story(id ID, snapshot provenance.Snapshot) (StoryResponse, error) {
	canonical, err := ParseID(id.String())
	if err != nil {
		return StoryResponse{}, err
	}
	metadata := graph.Metadata()
	story := StoryResponse{
		Schema: StorySchema, SemanticID: ID(canonical),
		SemanticDigest: metadata.SemanticDigest, GraphHash: metadata.GraphHash,
		Relationships: []Fact{}, Evidence: []StoryEvidenceReference{}, Status: StoryUnknown,
		Reason: "UNKNOWN_SEMANTIC_ID",
	}
	node, found := graph.Node(ID(canonical))
	if !found {
		return finalizeStory(story), nil
	}
	story.Node = node
	for _, fact := range graph.AllFacts() {
		if fact.Subject == ID(canonical) || fact.Object == ID(canonical) {
			story.Relationships = append(story.Relationships, fact)
		}
	}
	sortFacts(story.Relationships)
	for _, record := range snapshot.Records {
		if record.SemanticID != canonical.String() {
			continue
		}
		story.Evidence = append(story.Evidence, StoryEvidenceReference{
			ID: record.ID, SemanticID: record.SemanticID, Producer: record.Producer,
			Kind: record.Kind, Status: record.Status, SourceSpan: record.SourceSpan,
			SourceDigest: record.SourceDigest, SemanticDigest: record.SemanticDigest,
			GraphDigest: record.GraphDigest, Sequence: record.Sequence,
			Predecessor: copyDigestLink(record.Predecessor),
			Attributes:  copyAttributes(record.Attributes), Freshness: record.Freshness,
			Hash: record.Hash,
		})
	}
	sort.SliceStable(story.Evidence, func(left, right int) bool {
		first, second := story.Evidence[left], story.Evidence[right]
		if first.Sequence != second.Sequence {
			return first.Sequence < second.Sequence
		}
		if first.ID != second.ID {
			return first.ID < second.ID
		}
		return first.Hash < second.Hash
	})
	story.LedgerDigest = snapshot.Digest
	if len(story.Evidence) == 0 {
		story.Reason = "MISSING_EVIDENCE"
		return finalizeStory(story), nil
	}
	if snapshot.Digest == "" {
		story.Reason = "MISSING_LEDGER_DIGEST"
		return finalizeStory(story), nil
	}
	story.Status, story.Reason = storyEvidenceStatus(story.Evidence)
	return finalizeStory(story), nil
}

func storyEvidenceStatus(values []StoryEvidenceReference) (string, string) {
	verified, rejected, candidate := false, false, false
	for _, value := range values {
		switch value.Status {
		case provenance.StatusVerified:
			verified = true
		case provenance.StatusRejected, provenance.StatusFailed:
			rejected = true
		case provenance.StatusCandidate, provenance.StatusDeferred:
			candidate = true
		}
	}
	if verified && rejected {
		return StoryContradictory, "VERIFIED_AND_REJECTED_EVIDENCE"
	}
	if rejected {
		return StoryRejected, "EVIDENCE_REJECTED"
	}
	if verified {
		return StoryAvailable, "VERIFIED_EVIDENCE_BOUND"
	}
	if candidate {
		return StoryCandidate, "EVIDENCE_REQUIRES_ACCEPTANCE"
	}
	return StoryUnknown, "EVIDENCE_STATUS_UNKNOWN"
}

func finalizeStory(value StoryResponse) StoryResponse {
	value.BindingDigest = ""
	payload, _ := json.Marshal(value)
	value.BindingDigest = digestBytes(payload)
	return value
}

func (value StoryResponse) Validate() error {
	if value.Schema != StorySchema || value.SemanticID == "" || value.Status == "" || value.Reason == "" {
		return errors.New("story identity is invalid")
	}
	switch value.Status {
	case StoryAvailable, StoryCandidate, StoryRejected, StoryContradictory, StoryUnknown:
	default:
		return errors.New("story status is invalid")
	}
	if value.BindingDigest == "" {
		return errors.New("story binding digest is missing")
	}
	expected := finalizeStory(value)
	if expected.BindingDigest != value.BindingDigest {
		return errors.New("story binding digest is invalid")
	}
	return nil
}

func copyDigestLink(value *provenance.DigestLink) *provenance.DigestLink {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}

func copyAttributes(value map[string]string) map[string]string {
	if len(value) == 0 {
		return nil
	}
	result := make(map[string]string, len(value))
	maps.Copy(result, value)
	return result
}
