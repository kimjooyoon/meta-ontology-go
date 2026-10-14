package extractor

import (
	"os"
	"testing"
)

func TestFillCandidateOriginCapacity25363(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	targets := []string{"cmd/gooo#hasPackageOriginQueryOptions", "cmd/gooo#packageOriginQuery", "cmd/gooo#parsePackageExecuteArgs", "cmd/gooo#runBodyOrigins", "cmd/gooo#selectStarterFiles", "cmd/gooo#writeBodyFillCandidateOrigins", "cmd/gooo#writePackageOriginInput", "internal/bodycodegen#InspectSourceFillCandidateOrigins", "internal/bodycodegen#SourceFillCandidateOriginInputs", "internal/bodycodegen#sourceFillCandidateOriginRow", "internal/bodyexecution#NativeOriginQuery.hasSubjectSelection", "internal/bodyexecution#ValidateNativeOriginQuery", "internal/bodyexecution#nativeOriginPage.addField", "internal/bodyexecution#nativeOriginPage.fieldFileLocation", "internal/bodyexecution#nativeOriginPage.visitFields", "internal/bodyexecution#selectRecordedNativeOrigins", "internal/bodyexecution#validateOriginSubjectSelection"}
	requireNamedCapacities(t, root, targets)
	t.Log("candidate25363:all17 accepted-base changed production declarations;unchanged75-line rendered helper cap;no producer/model execution")
}
