package query

import provenance "github.com/kimjooyoon/meta-ontology-go/internal/provenance"

const StoryMismatchedBindingReason = "MISMATCHED_EVIDENCE_BINDING"

// StoryWithEvidenceBinding keeps a story read-only while refusing to promote
// evidence that is not bound to the current semantic and graph snapshot.
// SourceDigest is required because an execution record without source identity
// cannot establish which declaration was evaluated.
func (graph Graph) StoryWithEvidenceBinding(id ID, snapshot provenance.Snapshot) (StoryResponse, error) {
	story, err := graph.Story(id, snapshot)
	if err != nil {
		return StoryResponse{}, err
	}
	if len(story.Evidence) == 0 {
		return story, nil
	}
	metadata := graph.Metadata()
	for _, evidence := range story.Evidence {
		if evidence.SourceDigest == "" ||
			evidence.SemanticDigest == "" || evidence.SemanticDigest != metadata.SemanticDigest ||
			evidence.GraphDigest == "" || evidence.GraphDigest != metadata.GraphHash {
			story.Status = StoryUnknown
			story.Reason = StoryMismatchedBindingReason
			return finalizeStory(story), nil
		}
	}
	return story, nil
}
