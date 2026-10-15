package workspaceexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

// ProgramAccessFact exposes one visited source access occurrence. Ranges belong
// to the selected activity body identified by BodySHA256. This is a source
// occurrence, not an observed runtime read or an individual element identity.
type ProgramAccessFact struct {
	RootID     string `json:"root_id"`
	ActivityID string `json:"activity_id"`
	Kind       string `json:"kind"`
	Expression string `json:"expression"`
	BodySHA256 string `json:"body_sha256"`
	Start      int    `json:"start_byte"`
	End        int    `json:"end_byte"`
	Truncated  bool   `json:"snippet_truncated"`
}

// ProgramAccessFacts returns a detached scalar record set for Gooo consumers.
// Read Analysis, Issues, Truncated and AccessesTruncated with these facts.
// Accesses are collected for unfiltered roots and capped at 256 occurrences.
func ProgramAccessFacts(report ProgramValueOrigins) []ProgramAccessFact {
	return append([]ProgramAccessFact(nil), report.Accesses...)
}

func (w *valueTransferWalker) recordAccess(rootID string, a *valueTransferActivity, node bodycodegen.BodyValueOrigin) {
	if w.query.CandidateID != "" {
		return
	}
	switch node.Kind {
	case "field_access", "index_access", "slice_access", "access_base",
		"index_operand", "slice_low", "slice_high", "slice_max":
	default:
		return
	}
	if w.accesses == nil {
		w.accesses = map[string]bool{}
	}
	key := rootID + "\x00" + a.id + "\x00" + node.ID
	if w.accesses[key] {
		return
	}
	if len(w.out.Accesses) >= 256 {
		w.out.AccessesTruncated = true
		return
	}
	w.accesses[key] = true
	span := node.Decoded
	if span.Start < 0 || span.End < span.Start || span.End > len(a.body) {
		w.issue(a.id, "ACCESS_SOURCE_RANGE_UNAVAILABLE")
		return
	}
	snippet, cut := originQuerySnippet(a.body[span.Start:span.End])
	w.out.Accesses = append(w.out.Accesses, ProgramAccessFact{RootID: rootID, ActivityID: a.id,
		Kind: node.Kind, Expression: snippet, BodySHA256: a.graph.BodySHA256,
		Start: span.Start, End: span.End, Truncated: cut})
}
