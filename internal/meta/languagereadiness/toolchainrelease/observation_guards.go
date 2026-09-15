package toolchainrelease

func observeReceiptGuards(target Target, item PlatformEvidence, summary *Summary) {
	receipt := item.Receipt
	if receipt.Platform != target || receipt.ArchiveFormat != target.ArchiveFormat {
		summary.PlatformMismatches++
	}
	if receipt.Toolchain != ExpectedToolchain {
		summary.ToolchainMismatches++
	}
	if receipt.HeadSHA == "" {
		summary.HeadMismatches++
	}
	if receipt.Build.VCSModified {
		summary.DirtyBuilds++
	}
	if receipt.Build.VCSRevision != receipt.HeadSHA {
		summary.VCSRevisionMismatches++
	}
	if receipt.Binary.Builds != 2 || !receipt.Binary.ReplayEqual {
		summary.BinaryReplayMismatches++
	}
	if receipt.Archive.Builds != 2 || !receipt.Archive.ReplayEqual {
		summary.ArchiveReplayMismatches++
	}
	if receipt.Smoke.SchemaVersion != "gooo-version/v1" || receipt.Smoke.Language != "gooo" {
		summary.SmokeFailures++
	}
	if !item.ArchiveVerified {
		summary.ChecksumDrift++
	}
	if !item.ReceiptVerified || receipt.Schema != PlatformReceiptSchema {
		summary.ReceiptDigestFailures++
	}
	summary.RepositoryWrites += receipt.RepositoryWrites
	summary.MutationAuthorities += receipt.MutationAuthorities
}
