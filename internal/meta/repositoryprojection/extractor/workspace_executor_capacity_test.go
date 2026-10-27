package extractor

import "testing"

func TestScopedGoooExecutorRenderedCapacity(t *testing.T) {
	targets := []string{
		"internal/packageruntime/workspacesplit#Request.Validate",
		"internal/packageruntime/workspacesplit#DecodeRequests",
		"internal/meta/generation#BuildWithOperationInputs",
		"internal/meta/generation#BindWorkspaceSplitInputs",
		"internal/meta/generation#workspaceSplitInputFailures",
		"internal/meta/generation#validWorkspaceSplitIndicator",
		"scripts/meta-execution#executeGoooSplit",
		"scripts/meta-execution#materializeGoooSplit",
		"scripts/meta-execution#materializeFreshGoooSplit",
		"scripts/meta-execution#runGoooSplitCommand",
		"scripts/meta-execution#goooSplitProcessFailure",
		"scripts/meta-execution#applyAndVerifyGoooSplit",
		"scripts/meta-execution#verifyInstalledGoooSplit",
		"scripts/meta-execution#verifyGoooSplitNative",
		"scripts/meta-execution#readGoooSplitReceipt",
		"scripts/meta-execution#verifyGoooSplitProposal",
		"scripts/meta-execution#verifyGoooSplitOrigins",
		"scripts/meta-execution#verifyGoooSplitRuntime",
		"scripts/meta-execution#decodeGoooSplitRuntime",
		"scripts/meta-execution#finishGoooSplitMaterialization",
		"scripts/meta-execution#canonicalGoooSplitReceipt",
		"scripts/meta-execution#restoreGoooSplitSnapshot",
		"scripts/meta-execution#unpackGoooSplitSnapshot",
		"scripts/meta-execution#writeGoooSnapshotFile",
		"scripts/meta-execution#goooSnapshotMembers",
		"scripts/meta-execution#parseGoooSnapshotMembers",
		"scripts/meta-execution#parseGoooSnapshotMember",
		"scripts/meta-execution#readGoooSnapshotCommand",
		"scripts/meta-execution#goooSplitPlannerFixture",
		"scripts/meta-execution#TestGoooSplitExecutorUsesActualPublisherAndNativeOracles",
		"scripts/meta-execution#TestGoooSplitRuntimeRequiresExactFinitePopulation",
		"scripts/meta-execution#TestGoooSplitSnapshotPreservesCommittedInputAndRejectsArchiveProjection",
	}
	observed, err := collectNamedCapacityFiles("../../../..", targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
