package closure

import "fmt"

func validateVerification(in Input, program programDocument, verification verificationDocument) error {
	if verification.Schema != ProgramVerificationSchema ||
		verification.SubjectSHA != in.SubjectSHA ||
		verification.Status != StatusVerified {
		return fmt.Errorf("program verification identity or status is invalid")
	}
	if verification.RepositoryWorkspaceWrites || verification.PromotionAuthorized ||
			verification.BindingCount != canonicalBindingCount || verification.OperationCount != 8 ||
		verification.StepCount != 4 {
		return fmt.Errorf("program verification contract is invalid")
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
			return fmt.Errorf("program verification digest mismatch")
		}
	}
	if !digestPattern.MatchString(verification.Digest) {
		return fmt.Errorf("verification receipt digest is invalid")
	}
	return nil
}
