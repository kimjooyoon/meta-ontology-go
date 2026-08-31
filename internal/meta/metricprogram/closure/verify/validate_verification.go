package verify

import "fmt"

func validateVerification(program programDocument, verification verificationDocument) error {
	if verification.Schema != verificationSchema ||
		verification.SubjectSHA != program.SubjectSHA ||
		verification.Status != "VERIFIED" ||
		verification.RepositoryWorkspaceWrites || verification.PromotionAuthorized ||
		verification.BindingCount != canonicalBindingCount || verification.OperationCount != 9 ||
		verification.StepCount != 4 {
		return fmt.Errorf("verification contract is invalid")
	}
	pairs := [][2]string{
		{verification.ProgramDigest, program.Digest},
		{verification.StrategyDigest, program.StrategyDigest},
		{verification.RegistryDigest, program.RegistryDigest},
		{verification.SourceDigest, program.SourceDigest},
		{verification.SemanticDigest, program.SemanticDigest},
	}
	for _, pair := range pairs {
		if pair[0] != pair[1] || !digestPattern.MatchString(pair[0]) {
			return fmt.Errorf("verification digest mismatch")
		}
	}
	return nil
}
