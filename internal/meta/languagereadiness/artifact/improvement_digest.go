package artifact

func sealImprovement(receipt ImprovementArtifact) ImprovementArtifact {
	receipt.ArtifactDigest = ""
	receipt.ArtifactDigest = digestJSON(receipt)
	return receipt
}
