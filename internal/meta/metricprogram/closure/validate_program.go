package closure

import "fmt"

func validateProgram(in Input, program programDocument) error {
	if program.Schema != ProgramSchema || program.Repository != in.Repository ||
		program.SubjectSHA != in.SubjectSHA {
		return fmt.Errorf("program identity is not bound to the closure input")
	}
	if program.ExecutionPolicy != ProgramExecutionPolicy ||
		program.RepositoryWorkspaceWrites || program.PromotionAuthorized {
		return fmt.Errorf("program is not read-only")
	}
	root := program.RootPolicy
	if root.CountsApplicability != "OBSERVED" ||
		root.TopologyApplicability != "NOT_APPLICABLE" ||
		root.TopologyReason != "ROOT_TOPOLOGY_EXEMPT" ||
		root.ReadmeRequirement != "NOT_APPLICABLE" {
		return fmt.Errorf("program root policy is not the project-root exception")
	}
	for _, digest := range []string{program.StrategyDigest, program.StrategyVerificationDigest,
		program.RegistryDigest, program.SourceDigest, program.SemanticDigest, program.Digest} {
		if !digestPattern.MatchString(digest) {
			return fmt.Errorf("program contains invalid digest %q", digest)
		}
	}
	return validateProgramMembers(program)
}
