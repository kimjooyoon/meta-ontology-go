package verify

func reconstruct(in Input, program programDocument, verification verificationDocument) (receipt, error) {
	value := receipt{
		Schema: closureSchema, Repository: in.Repository, SubjectSHA: in.SubjectSHA,
		RunID: in.RunID, RunAttempt: in.RunAttempt, ExecutionPolicy: closurePolicy,
		RootPolicy: program.RootPolicy, Artifact: in.Artifact,
		Files: files{
			Program:      fileEvidence{Path: "program.json", Digest: digestBytes(in.ProgramJSON)},
			Source:       fileEvidence{Path: "program.gooo", Digest: digestBytes(in.Source)},
			Verification: fileEvidence{Path: "verification.json", Digest: digestBytes(in.VerificationJSON)},
		},
		ProgramDigest: program.Digest, VerificationDigest: verification.Digest,
		StrategyDigest: program.StrategyDigest, RegistryDigest: program.RegistryDigest,
		SourceDigest: program.SourceDigest, SemanticDigest: program.SemanticDigest,
		BindingCount: canonicalBindingCount, OperationCount: 9, StepCount: 4,
		Status: "VERIFIED", WriteEffect: "none",
		RepositoryWorkspaceWrites: false, PromotionAuthorized: false,
	}
	var err error
	value.Indicators, err = expectedIndicators(value)
	if err != nil {
		return receipt{}, err
	}
	value.Digest, err = receiptDigest(value)
	return value, err
}
