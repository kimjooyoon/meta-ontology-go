package toolchainrelease

func observeTarget(target Target, items []PlatformEvidence, summary *Summary,
	observations map[string]caseObservation) {
	if len(items) == 0 {
		summary.MissingReceipts++
		return
	}
	if len(items) > 1 {
		summary.DuplicateReceipts += len(items) - 1
	}
	item := items[0]
	receipt := item.Receipt
	summary.PlatformReceipts++
	observeReceiptGuards(target, item, summary)
	addDriverCounts(receipt, item, summary)
	prefix := target.ID + "-"
	observations[prefix+"receipt"] = caseObservation{"EXACT_RECEIPT", receipt.ReceiptDigest, item.ReceiptVerified}
	platformReady := receipt.Platform == target
	observations[prefix+"native-runtime"] = caseObservation{receipt.Platform.GOOS + "/" + receipt.Platform.GOARCH, receipt.ReceiptDigest, platformReady}
	toolchainReady := receipt.Toolchain == ExpectedToolchain
	observations[prefix+"go127-toolchain"] = caseObservation{receipt.Toolchain, receipt.ReceiptDigest, toolchainReady}
	binaryReady := receipt.Binary.Builds == 2 && receipt.Binary.ReplayEqual
	observations[prefix+"binary-replay"] = caseObservation{"2_BYTE_EQUAL_BUILDS", receipt.Binary.Digest, binaryReady}
	archiveReady := receipt.Archive.Builds == 2 && receipt.Archive.ReplayEqual && item.ArchiveVerified
	observations[prefix+"archive-replay"] = caseObservation{"2_BYTE_EQUAL_ARCHIVES", receipt.Archive.Digest, archiveReady}
	smokeReady := receipt.Smoke.SchemaVersion == "gooo-version/v1" && receipt.Smoke.Language == "gooo"
	observations[prefix+"version-smoke"] = caseObservation{receipt.Smoke.SchemaVersion, receipt.ReceiptDigest, smokeReady}
}

func addDriverCounts(receipt PlatformReceipt, item PlatformEvidence, summary *Summary) {
	summary.BinaryBuilds += receipt.Binary.Builds
	summary.ArchiveBuilds += receipt.Archive.Builds
	if receipt.Smoke.SchemaVersion == "gooo-version/v1" {
		summary.NativeSmokes++
	}
	if receipt.Binary.ReplayEqual {
		summary.BinaryReplays++
	}
	if receipt.Archive.ReplayEqual {
		summary.ArchiveReplays++
	}
	if item.ArchiveVerified {
		summary.ChecksumEntries++
	}
	if receipt.Toolchain == ExpectedToolchain {
		summary.ToolchainBindings++
	}
	if !receipt.Build.VCSModified && receipt.Build.VCSRevision == receipt.HeadSHA {
		summary.VCSBindings++
	}
}
