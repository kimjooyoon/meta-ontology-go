package toolchainrelease

func guardrailTotal(s Summary) int {
	return s.MissingReceipts + s.DuplicateReceipts + s.UnexpectedReceipts +
		s.CaseFailures + s.PlatformMismatches + s.ToolchainMismatches +
		s.HeadMismatches + s.DirtyBuilds + s.VCSRevisionMismatches +
		s.BinaryReplayMismatches + s.ArchiveReplayMismatches + s.SmokeFailures +
		s.ChecksumDrift + s.ReceiptDigestFailures + s.CorpusDrift +
		s.ConceptDrift + s.ProofFailures + s.Unresolved +
		s.RepositoryWrites + s.MutationAuthorities
}
