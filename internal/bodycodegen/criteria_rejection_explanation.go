package bodycodegen

import "encoding/json"

// This is derived from recorded question history. It does not run a candidate,
// score a predicate, or claim that the criteria are inconsistent for other code.
type CriteriaRejectionExplanation struct {
	Schema           string                   `json:"schema"`
	Scope            string                   `json:"scope"`
	Status           string                   `json:"status"`
	ReceiptSHA256    string                   `json:"receipt_sha256"`
	SourceSHA256     string                   `json:"source_sha256"`
	CandidateSpace   int                      `json:"candidate_space"`
	Considered       int                      `json:"considered"`
	Unobserved       int                      `json:"unobserved"`
	Surviving        int                      `json:"surviving"`
	Unavailable      int                      `json:"unavailable"`
	Rejected         int                      `json:"rejected"`
	Explained        int                      `json:"explained"`
	UnexplainedMasks []uint16                 `json:"unexplained_masks"`
	Groups           []CriteriaRejectionGroup `json:"groups"`
}

type CriteriaRejectionGroup struct {
	Kind         string                     `json:"kind"`
	Criterion    *RecordCriterionBinding    `json:"criterion,omitempty"`
	InputJSON    string                     `json:"input_json"`
	ExpectedJSON string                     `json:"expected_json,omitempty"`
	Witnesses    []CriteriaRejectionWitness `json:"witnesses"`
}

type CriteriaRejectionWitness struct {
	Mask            uint16 `json:"mask"`
	ActualJSON      string `json:"actual_json"`
	OutputIndex     int    `json:"output_index"`
	EvaluationIndex *int   `json:"evaluation_index,omitempty"`
}

func ExplainCriteriaRejections(a *RecordAssemblyReceipt) CriteriaRejectionExplanation {
	r := CriteriaRejectionExplanation{Schema: "gooo/criteria-rejection-explanation/v1", Status: "NO_QUESTION_HISTORY",
		Scope: "initial recorded question history; one linked decisive failure per rejected candidate; other implementations may satisfy the conditions"}
	if a == nil || a.CriteriaQuestions == nil || len(a.CriteriaQuestions.Rounds) == 0 {
		return r
	}
	q := a.CriteriaQuestions
	r.Status, r.SourceSHA256 = "INCOMPLETE_OBSERVATION", q.SourceSHA256
	if (q.Schema != "gooo/criteria-questions/v1" && q.Schema != "gooo/criteria-questions/v2") || q.Status == "STARTED" || q.Status == "FAILED" {
		return r
	}
	raw, err := json.Marshal(a)
	if err != nil || len(a.Choices) < 1 || len(a.Choices) > 6 {
		return r
	}
	r.ReceiptSHA256, r.CandidateSpace = digest(raw), 1<<len(a.Choices)
	last := q.Rounds[len(q.Rounds)-1]
	if !consistentCriteriaRound(last, r.CandidateSpace) {
		return r
	}
	r.Considered, r.Unobserved = last.Candidates, last.Unobserved
	r.Surviving, r.Unavailable, r.Rejected = len(last.SurvivingMasks), len(last.UnavailableMasks), len(last.RejectedMasks)
	for _, candidate := range last.Evaluations {
		if candidate.Status != "UNSATISFIED" {
			continue
		}
		group, ok := criteriaRejectionWitness(a, candidate)
		if !ok {
			r.UnexplainedMasks = append(r.UnexplainedMasks, candidate.Mask)
			continue
		}
		r.Explained++
		r.Groups = appendCriteriaRejection(r.Groups, group)
	}
	r.Status = "CANDIDATES_REMAIN"
	if r.Explained != r.Rejected {
		r.Status = "INCOMPLETE_REJECTION_EVIDENCE"
	} else if r.Rejected == r.Considered {
		r.Status = "EXAMINED_CANDIDATES_REJECTED"
		if r.Unobserved == 0 {
			r.Status = "ALL_DECLARED_CANDIDATES_REJECTED"
		}
	}
	return r
}

func appendCriteriaRejection(groups []CriteriaRejectionGroup, added CriteriaRejectionGroup) []CriteriaRejectionGroup {
	key := criteriaRejectionGroupKey(added)
	for i := range groups {
		if criteriaRejectionGroupKey(groups[i]) == key {
			groups[i].Witnesses = append(groups[i].Witnesses, added.Witnesses...)
			return groups
		}
	}
	return append(groups, added)
}

func criteriaRejectionGroupKey(group CriteriaRejectionGroup) string {
	group.Witnesses = nil
	raw, _ := json.Marshal(group)
	return string(raw)
}
