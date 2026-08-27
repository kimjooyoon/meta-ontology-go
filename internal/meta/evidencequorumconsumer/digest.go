package evidencequorumconsumer

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumwire"

func reportDigest(report Report) string {
	digest := report.Digest
	report.Digest = ""
	return evidencequorumwire.DigestJSON(reportWithoutDigest{Report: report})
}

// A named wrapper keeps the digest input stable if Report gains methods or
// embedding in a later schema revision.
type reportWithoutDigest struct {
	Report
}

func previousClaimDigest(claimID string) string {
	return evidencequorumwire.DigestJSON(struct {
		State   string `json:"state"`
		ClaimID string `json:"claim_id"`
	}{State: "OPEN", ClaimID: claimID})
}
