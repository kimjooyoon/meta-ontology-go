package guardedpromotion

func validPromotionArtifact(source Source) bool {
	artifact := source.Artifact
	expectedName := PromotionArtifactBase + source.PredecessorSHA
	return artifact.ArtifactName == expectedName &&
		validDigest(artifact.ArtifactDigest) &&
		validDigest(artifact.FileSHA256) &&
		(artifact.RunEvent == "push" || artifact.RunEvent == "workflow_run") &&
		artifact.ReportSchema == PromotionSchema &&
		validDigest(artifact.ReportDigest) &&
		artifact.ReportCurrentHeadSHA == source.PredecessorSHA &&
		artifact.ReportDecision == "PASS" &&
		artifact.ReportSatisfied == 8 &&
		artifact.ReportTotal == 8 &&
		artifact.ReportUnresolved == 0 &&
		artifact.ReportRepositoryWrites == 0
}
