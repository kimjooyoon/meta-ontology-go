package extractor

import (
	"os"
	"testing"
)

func TestRegisteredWorkspaceRenderedCapacity(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	targets := []string{
		"cmd/gooo#prepareRegisteredWorkspaceSplit",
		"cmd/gooo#splitRegistrationLocation",
		"cmd/gooo#workspaceRegistrationPartition",
		"cmd/gooo#TestWorkspaceRegisteredSplitRestoresJointWrites",
		"cmd/gooo#TestWorkspaceRegisteredSplitRefusesChangedInputsAndSavedOutputs",
		"cmd/gooo#TestWorkspaceRegisteredSplitRechecksUnchangedNativeAfterStaging",
		"cmd/gooo#TestWorkspaceRegisteredSplitAddsExternalManifestOnlyOnce",
		"cmd/gooo#splitRegistrationTarget",
		"cmd/gooo#splitRegistrationEmbedInputs",
		"cmd/gooo#captureSplitAncestorGo",
		"cmd/gooo#requireUnembeddedSplit",
		"cmd/gooo#TestWorkspaceRegisteredSplitRequiresNativeAdmissionForEmbeddedConsumers",
		"cmd/gooo#TestWorkspaceRegisteredSplitFlagCombinationsAndRoot",
		"cmd/gooo#TestWorkspaceRegisteredSplitAncestorReadFailure",
		"cmd/gooo#registeredSplitFixture",
		"cmd/gooo#copyRegisteredSplitInput",
		"cmd/gooo#writeRegisteredSplitInput",
		"cmd/gooo#saveRegisteredSplit",
		"cmd/gooo#TestWorkspaceRegisteredSplitCLIInstallsAndExecutes",
		"cmd/gooo#describeSplitRegistration",
		"cmd/gooo#captureSplitOperationInputs",
		"cmd/gooo#appendSplitRegistrationWrites",
		"cmd/gooo#splitRegistrationWrite",
		"cmd/gooo#identicalSplitWrite",
		"cmd/gooo#verifySplitRegistrationOutput",
		"internal/meta/generation#SplitGoooInputContractEvidence",
		"internal/meta/generation#TestSplitGoooOperationRetainsSourceActivityIdentity",
		"internal/packageruntime#ChooseSourceSplit",
		"internal/packageruntime#chooseSourceSplit",
		"internal/packageruntime#TestSourceSplitChoiceMatchesMaterializedOutputs",
		"internal/packageruntime#TestSourceSplitChoiceCountsPartialProgressAndStableTies",
		"cmd/gooo#parseSplitLineLimit",
		"cmd/gooo#prepareChosenWorkspaceSplit",
		"cmd/gooo#chosenRegisteredSplit",
		"cmd/gooo#TestWorkspaceChosenSplitPublishesAndExecutes",
		"cmd/gooo#assertChosenSplitRuntime",
		"cmd/gooo#TestWorkspaceSplitChoiceModesAreExclusive",
		"cmd/gooo#TestWorkspaceSplitChoiceRecomputedBeforePublication",
		"cmd/gooo#TestWorkspaceSplitChoiceRetainsSatisfiedSource",
		"cmd/gooo#changingSplitChoiceReader.ReadFile",
		"cmd/gooo#TestWorkspaceSplitChoiceRejectsChangingSelectionInputs",
		"cmd/gooo#runWorkspaceSourceSplit",
		"cmd/gooo#parseWorkspaceSourceSplit",
		"cmd/gooo#validateWorkspaceSourceSplitOptions",
		"cmd/gooo#prepareWorkspaceSourceSplit",
		"cmd/gooo#applyWorkspaceSourceSplit",
	}
	observed, err := collectNamedCapacityFiles(root, targets, true)
	if err != nil {
		t.Fatal(err)
	}
	requireObservedCapacities(t, targets, observed)
}
