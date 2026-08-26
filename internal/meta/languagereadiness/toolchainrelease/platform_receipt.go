package toolchainrelease

func newPlatformReceipt(input BuildInput, build BuildEvidence, smoke SmokeEvidence,
	archive ReplayEvidence, binaryDigest string, binaryBytes int64) PlatformReceipt {
	return PlatformReceipt{
		Schema:     PlatformReceiptSchema,
		Decision:   DecisionPass,
		Reason:     "TOOLCHAIN_RELEASE_PLATFORM_READY",
		Resolution: ResolutionExact,
		HeadSHA:    input.ExpectedHead,
		Toolchain:  ExpectedToolchain,
		Platform:   input.Target,
		Build:      build,
		Binary: ReplayEvidence{
			Name: binaryName(input.Target), Digest: binaryDigest,
			Bytes: binaryBytes, Builds: 2, ReplayEqual: true,
		},
		Archive:             archive,
		ArchiveFormat:       input.Target.ArchiveFormat,
		Smoke:               smoke,
		RepositoryWrites:    0,
		MutationAuthorities: 0,
	}
}
