package guardedpromotion

func Coordinates(source Source) []Coordinate {
	candidateKnown := source.CollectionError == "" && source.UnresolvedCandidates == 0
	candidateOK := source.ValidCandidates == 1 && source.AmbiguousCandidates == 0
	artifactKnown := candidateKnown && candidateOK && source.Artifact.ArtifactID > 0
	workflowName, eventKnown := expectedCIName(source.Workflow.Event)
	return []Coordinate{
		resolvedCoordinate("repository-identity", "FOUNDATION",
			source.RequestedRepository != "" && source.ObservedRepository != "",
			source.RequestedRepository == source.ObservedRepository),
		resolvedCoordinate("current-subject-sha", "FOUNDATION",
			validSHA(source.CurrentHeadSHA), validSHA(source.CurrentHeadSHA)),
		resolvedCoordinate("predecessor-sha", "FOUNDATION",
			validSHA(source.PredecessorSHA), validSHA(source.PredecessorSHA) &&
				source.PredecessorSHA != source.CurrentHeadSHA),
		resolvedCoordinate("unique-predecessor-artifact", "FOUNDATION",
			candidateKnown, candidateOK),
		resolvedCoordinate("predecessor-promotion-contract", "FOUNDATION",
			artifactKnown, validPromotionArtifact(source)),
		resolvedCoordinate("ci-workflow-identity", "COHERENCE",
			eventKnown && source.Workflow.Name != "" && source.Workflow.Path != "",
			source.Workflow.Name == workflowName && source.Workflow.Path == CIPath),
		resolvedCoordinate("ci-workflow-success", "COHERENCE",
			source.Workflow.Status != "" && source.Workflow.Conclusion != "",
			source.Workflow.Status == "completed" && source.Workflow.Conclusion == "success"),
		resolvedCoordinate("ci-subject-link", "COHERENCE",
			validSHA(source.Workflow.HeadSHA), source.Workflow.HeadSHA == source.CurrentHeadSHA),
		resolvedCoordinate("merged-push-event", "REGRESSION",
			eventKnown, source.Workflow.Event == "push"),
		resolvedCoordinate("default-branch-boundary", "REGRESSION",
			source.DefaultBranch != "" && source.Workflow.HeadBranch != "",
			source.Workflow.HeadBranch == source.DefaultBranch),
		resolvedCoordinate("observer-write-boundary", "REGRESSION",
			true, source.RepositoryWrites == 0),
		resolvedCoordinate("mutation-authority-boundary", "REGRESSION",
			true, !source.RepositoryMutationAuthorized),
	}
}
