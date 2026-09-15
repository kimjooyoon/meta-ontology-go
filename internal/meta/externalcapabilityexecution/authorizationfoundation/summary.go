package authorizationfoundation

import "fmt"

func SummaryMarkdown(receipt Receipt, suite Suite) []byte {
	foundation := receipt.Foundation
	return []byte(fmt.Sprintf(`### External capability authorization foundation

- decision: %s / %s / %s
- CAB-10: %d/%d (%d bps); unknown/open/rejected: %d/%d/%d
- claims: discharged %d/%d
- conformance: %d/%d; authorized/unknown/denied: %d/%d/%d
- foundation artifact: %d (%s)
- authority: execution=%t mutation=%t promotion=%t; writes=%d
`, receipt.Decision, receipt.Resolution, receipt.EnforcementEffect,
		receipt.Completed, receipt.Total, receipt.BasisPoints,
		receipt.UnknownIndicators, receipt.OpenClaims, receipt.RejectedClaims,
		receipt.DischargedClaims, receipt.Total, suite.Passed, suite.Total,
		suite.AuthorizedCases, suite.UnknownCases, suite.DeniedCases,
		foundation.ArtifactID, foundation.ArchiveDigest,
		receipt.ExecutionAuthority, receipt.RepositoryMutationAuthority,
		receipt.PromotionAuthority, receipt.RepositoryWrites))
}
