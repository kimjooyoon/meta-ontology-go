package feedbackpredecessor

type FoundationEvidence struct {
	ProofChoice                 string   `json:"proof_choice"`
	Reason                      string   `json:"reason"`
	LastKnownGoodSHA            string   `json:"last_known_good_sha"`
	LastKnownGoodRunID          int64    `json:"last_known_good_run_id"`
	LastKnownGoodArtifactID     int64    `json:"last_known_good_artifact_id"`
	LastKnownGoodArtifactName   string   `json:"last_known_good_artifact_name"`
	LastKnownGoodArtifactDigest string   `json:"last_known_good_artifact_digest"`
	LastKnownGoodReceiptDigest  string   `json:"last_known_good_receipt_digest"`
	LastKnownGoodIsAncestor     bool     `json:"last_known_good_is_ancestor"`
	MissingPredecessorSHA       string   `json:"missing_predecessor_sha"`
	GapCommitSHAs               []string `json:"gap_commit_shas"`
	GapPRNumbers                []int    `json:"gap_pr_numbers"`
	NextOperation               string   `json:"next_operation"`
	UseCount                    int      `json:"use_count"`
	BlockedBy                   []string `json:"blocked_by"`
}

func FoundationEvidenceForConfirmedGap() FoundationEvidence {
	return FoundationEvidence{
		ProofChoice:                 FoundationProofChoice,
		Reason:                      ReasonFoundationRegression,
		LastKnownGoodSHA:            FoundationLastKnownGoodSHA,
		LastKnownGoodRunID:          FoundationLastKnownGoodRunID,
		LastKnownGoodArtifactID:     FoundationLastKnownGoodArtifactID,
		LastKnownGoodArtifactName:   FoundationLastKnownGoodArtifactName,
		LastKnownGoodArtifactDigest: FoundationLastKnownGoodArtifactDigest,
		LastKnownGoodReceiptDigest:  FoundationLastKnownGoodReceiptDigest,
		LastKnownGoodIsAncestor:     true,
		MissingPredecessorSHA:       FoundationMissingPredecessorSHA,
		GapCommitSHAs: []string{
			"cc4a47ec8059e385fabd709a42894ad607ae16c7",
			"9905abfdaea296af03f5b265da50128083caed53",
			"cd9727af80f5118405290d3be96890c18e1529c0",
		},
		GapPRNumbers:  []int{599, 600, 601},
		NextOperation: FoundationNextOperation,
		UseCount:      1,
		BlockedBy: []string{
			"PR #599: storage.direct-entry=1",
			"PR #600: workflow discovery-root classification incomplete",
			"PR #601: FLOOR_REGRESSION",
			"PREDECESSOR_FEEDBACK_NOT_FOUND",
		},
	}
}
