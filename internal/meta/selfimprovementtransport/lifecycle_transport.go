package selfimprovementtransport

func CompleteArtifactLifecycle(receipt LifecycleReceipt, archiveRaw []byte, downloadExit int) LifecycleReceipt {
	if len(receipt.Indicators) != lifecycleFixedStepTotal ||
		receipt.Indicators[0].Status != StatusVerified ||
		receipt.Indicators[1].Status != StatusVerified ||
		receipt.Indicators[2].Status != StatusVerified {
		closeLifecycle(&receipt)
		return receipt
	}
	if downloadExit != 0 || len(archiveRaw) == 0 {
		failLifecycleAt(&receipt, 3, "ARTIFACT_DOWNLOAD_FAILED", "", "", "")
		closeLifecycle(&receipt)
		return receipt
	}
	actual := digestBytes(archiveRaw)
	receipt.ActualArchiveDigest = actual
	verifyLifecycleStep(&receipt, 3, actual)
	if actual != receipt.ArtifactDigest {
		failLifecycleAt(&receipt, 4, "ARCHIVE_DIGEST_MISMATCH",
			receipt.ArtifactDigest, actual, "CONTRADICTION")
		closeLifecycle(&receipt)
		return receipt
	}
	verifyLifecycleStep(&receipt, 4, actual)
	closeLifecycle(&receipt)
	return receipt
}
