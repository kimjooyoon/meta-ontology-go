package closure

func newReceipt(in Input, program programDocument, verification verificationDocument) Receipt {
	return Receipt{
		Schema: Schema, Repository: in.Repository, SubjectSHA: in.SubjectSHA,
		RunID: in.RunID, RunAttempt: in.RunAttempt, ExecutionPolicy: ExecutionPolicy,
		RootPolicy: program.RootPolicy, Artifact: in.Artifact,
		Files: Files{
			Program:      FileEvidence{Path: "program.json", Digest: digestBytes(in.ProgramJSON)},
			Source:       FileEvidence{Path: "program.gooo", Digest: digestBytes(in.Source)},
			Verification: FileEvidence{Path: "verification.json", Digest: digestBytes(in.VerificationJSON)},
		},
		ProgramDigest: program.Digest, VerificationDigest: verification.Digest,
		StrategyDigest: program.StrategyDigest, RegistryDigest: program.RegistryDigest,
		SourceDigest: program.SourceDigest, SemanticDigest: program.SemanticDigest,
			BindingCount: canonicalBindingCount, OperationCount: 8, StepCount: 4,
		Status: StatusVerified, WriteEffect: WriteEffectNone,
		RepositoryWorkspaceWrites: false, PromotionAuthorized: false,
	}
}
