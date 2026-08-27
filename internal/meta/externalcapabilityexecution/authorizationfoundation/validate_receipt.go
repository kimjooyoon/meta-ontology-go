package authorizationfoundation

func validateBootstrap(value Receipt, subject string, foundation Foundation, prior bool) error {
	identity := value.Schema == BootstrapSchema && value.SubjectSHA == subject &&
		value.Decision == "FAIL_CLOSED" && value.Resolution == "UNKNOWN" &&
		value.EnforcementEffect == "BLOCK" && value.Reason == "CAPABILITY_AUTHORIZATION_EVIDENCE_UNKNOWN"
	counts := value.Completed == 9 && value.Total == 10 && value.BasisPoints == 9000 &&
		value.UnknownIndicators == 1 && value.OpenClaims == 1 &&
		value.DischargedClaims == 9 && value.RejectedClaims == 0
	if !identity || !counts || len(value.Indicators) != 10 || len(value.Claims) != 10 || len(value.Unknowns) != 1 {
		return denied("BOOTSTRAP_AUTHORIZATION_CONTRACT_MISMATCH")
	}
	if value.PolicySourceDigest != foundation.PolicySourceDigest ||
		value.PolicyGeneratedDigest != foundation.PolicyGeneratedDigest {
		return denied("BOOTSTRAP_POLICY_DIGEST_MISMATCH")
	}
	if prior && value.ReceiptDigest != foundation.ReceiptDigest {
		return denied("BOOTSTRAP_RECEIPT_DIGEST_MISMATCH")
	}
	if value.RepositoryWrites != 0 || value.OfficialMutationCount != 0 || value.PromotionCount != 0 ||
		value.ExecutionAuthority || value.RepositoryMutationAuthority || value.PromotionAuthority {
		return denied("BOOTSTRAP_AUTHORITY_CEILING_BREACHED")
	}
	return validateBootstrapComponents(value)
}
