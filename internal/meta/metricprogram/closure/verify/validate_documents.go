package verify

import "fmt"

func validateDocuments(in Input, program programDocument, verification verificationDocument) error {
	if program.Schema != programSchema || program.Repository != in.Repository ||
		program.SubjectSHA != in.SubjectSHA || program.ExecutionPolicy != programPolicy {
		return fmt.Errorf("program identity or execution policy is invalid")
	}
	if program.RepositoryWorkspaceWrites || program.PromotionAuthorized ||
		len(program.Operations) != 8 || len(program.Bindings) != 15 || len(program.Steps) != 4 {
		return fmt.Errorf("program cardinality or write policy is invalid")
	}
	root := program.RootPolicy
	if root.CountsApplicability != "OBSERVED" ||
		root.TopologyApplicability != "NOT_APPLICABLE" ||
		root.TopologyReason != "ROOT_TOPOLOGY_EXEMPT" ||
		root.ReadmeRequirement != "NOT_APPLICABLE" {
		return fmt.Errorf("project-root exception is invalid")
	}
	if digestBytes(in.Source) != program.SourceDigest {
		return fmt.Errorf("source digest mismatch")
	}
	return validateFixedPoint(program, verification)
}
