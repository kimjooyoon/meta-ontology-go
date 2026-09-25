package selfimprovementcandidate

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

func buildProofs(report Report, success bool) []Proof {
	foundation := report.SourceObservationDigest
	if !validDigest(foundation) {
		foundation = report.SourceFileDigest
	}
	coherence := report.SourceFileDigest
	if len(report.Candidates) == 1 {
		coherence = report.Candidates[0].Digest
	}
	return []Proof{
		{Choice: "FOUNDATION", Claim: "the exact observation and Gooo candidate contract are fixed",
			MetaOperation: "bind-nonexecuting-candidate-foundation", EvidenceDigest: foundation, Passed: success},
		{Choice: "COHERENCE", Claim: "the explicit nonclaim agrees with the numeric experiment target",
			MetaOperation: "compare-gap-and-candidate", EvidenceDigest: coherence, Passed: success},
		{Choice: "REGRESSION", Claim: "candidate generation grants no effects or authority",
			MetaOperation: "guard-nonexecuting-candidate", EvidenceDigest: digestJSON(report.Authority), Passed: success},
	}
}
