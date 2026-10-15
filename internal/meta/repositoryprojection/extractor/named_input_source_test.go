package extractor

import "testing"

// Declaration identity survives repository source partitioning. The original
// 60-file extraction observation remains in docs/observations/named-inputs.
func TestNamedInputSourcePartitionsFirst(t *testing.T) {
	requireNamedCapacities(t, "../../../..", []string{
		"internal/bidir#bindSemanticInputs",
		"internal/bidir#bindSemanticFields",
		"internal/bodycodegen#renameLoadedActivity",
		"internal/bodycodegen#collectActivityRenameReferences",
		"internal/bodycodegen#finishActivityRename",
		"internal/packageruntime/workspaceexecution#selectedValueActivityNames",
		"internal/packageruntime/workspaceexecution#addSelectedValueActivities",
		"internal/packageruntime/workspaceexecution#bindSelectedValueActivities",
		"internal/packageruntime/workspaceexecution#retainSelectedValueFragments",
		"internal/packageruntime/workspaceexecution#valueTransferWalker.walkValueVisit",
		"internal/packageruntime/workspaceexecution#valueTransferWalker.emitValueCandidates",
		"internal/packageruntime/workspaceexecution#valueTransferWalker.followValueCall",
		"internal/packageruntime/workspaceexecution#valueTransferWalker.followValueInput",
		"internal/syntax#Parser.parseActivityInputs",
		"internal/inputbinding#Resolve",
		"internal/inputbinding#PositionName",
	})
}
