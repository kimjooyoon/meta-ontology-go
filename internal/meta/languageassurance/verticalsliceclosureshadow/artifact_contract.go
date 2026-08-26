package verticalsliceclosureshadow

func validArtifactSubject(id string, artifact artifactEnvelope, head string) bool {
	if !validSHA(head) {
		return false
	}
	if id == "release" {
		return artifact.HeadSHA == head
	}
	return artifact.Source.ExpectedHeadSHA == head &&
		(artifact.HeadSHA == "" || artifact.HeadSHA == head)
}

func observedMetaOperation(id string, artifact artifactEnvelope) string {
	if artifact.MetaOperation != "" {
		return artifact.MetaOperation
	}
	if artifact.Source.MetaOperation != "" {
		return artifact.Source.MetaOperation
	}
	return codeMetaOperation(id)
}

func mutationAuthorityPresent(id string, artifact artifactEnvelope) bool {
	if artifact.MutationAuthorized != nil && *artifact.MutationAuthorized {
		return true
	}
	return id != "release" && artifact.MutationAuthorized == nil
}
