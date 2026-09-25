package authorizationfoundation

import "fmt"

func Validate(value Receipt, subject string) error {
	exact := value.Schema == ReceiptSchema && value.SubjectSHA == subject &&
		value.Decision == "AUTHORIZED_SHADOW" && value.Resolution == "EXACT" &&
		value.EnforcementEffect == "NO_EFFECT" && value.Reason == "CAPABILITY_AUTHORIZATION_FOUNDATION_BOUND"
	counts := value.Completed == 10 && value.Total == 10 && value.BasisPoints == 10000 &&
		value.UnknownIndicators == 0 && value.OpenClaims == 0 &&
		value.DischargedClaims == 10 && value.RejectedClaims == 0 && len(value.Unknowns) == 0
	ceiling := value.RepositoryWrites == 0 && value.OfficialMutationCount == 0 && value.PromotionCount == 0 &&
		!value.ExecutionAuthority && !value.RepositoryMutationAuthority && !value.PromotionAuthority
	if !exact || !counts || !ceiling || value.Foundation == nil {
		return fmt.Errorf("closed authorization receipt is not exact")
	}
	if value.Foundation.ArtifactID != ExpectedArtifactID ||
		value.Foundation.ArchiveDigest != ExpectedArchiveDigest || value.Foundation.EvidenceDigest == "" {
		return fmt.Errorf("closed authorization foundation is not exact")
	}
	expectedDigest := value.ReceiptDigest
	if sealReceipt(value).ReceiptDigest != expectedDigest {
		return fmt.Errorf("closed authorization receipt digest mismatch")
	}
	return nil
}
