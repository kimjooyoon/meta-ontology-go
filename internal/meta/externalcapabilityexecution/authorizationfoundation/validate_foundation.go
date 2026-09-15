package authorizationfoundation

func validateFoundation(value Foundation) error {
	exact := value.Schema == FoundationSchema && value.Repository == ExpectedRepository &&
		value.ProducerRunID == ExpectedRunID && value.ProducerRunAttempt == ExpectedRunAttempt &&
		value.SubjectSHA == ExpectedSubjectSHA && value.ArtifactID == ExpectedArtifactID &&
		value.ArtifactName == ExpectedArtifactName && value.ArchiveDigest == ExpectedArchiveDigest &&
		value.ReceiptFileDigest == ExpectedFileDigest && value.ReceiptDigest == ExpectedReceiptDigest &&
		value.PolicySourceDigest == ExpectedSourceDigest && value.PolicyGeneratedDigest == ExpectedTreeDigest
	if !exact {
		return denied("POLICY_FOUNDATION_CONTRACT_MISMATCH")
	}
	bootstrap := value.BootstrapDecision == "FAIL_CLOSED" && value.BootstrapResolution == "UNKNOWN" &&
		value.BootstrapUnknownStage == PolicyStage &&
		value.BootstrapUnknownReason == "POLICY_FOUNDATION_UNAVAILABLE"
	if !bootstrap || value.RepositoryWrites != 0 || value.MutationAuthority || value.PromotionAuthority {
		return denied("POLICY_FOUNDATION_BOOTSTRAP_MISMATCH")
	}
	return nil
}

func validateMetadata(value ArtifactMetadata, foundation Foundation) error {
	if value.Expired {
		return unknown("POLICY_FOUNDATION_ARTIFACT_EXPIRED")
	}
	if value.ID != foundation.ArtifactID || value.Name != foundation.ArtifactName ||
		value.Digest != foundation.ArchiveDigest || value.CreatedAt == "" || value.ExpiresAt == "" {
		return denied("POLICY_FOUNDATION_ARTIFACT_MISMATCH")
	}
	return nil
}
