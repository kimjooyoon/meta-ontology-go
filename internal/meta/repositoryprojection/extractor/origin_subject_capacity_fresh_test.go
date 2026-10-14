package extractor

import (
	"os"
	"testing"
)

func TestOriginSubjectCapacity24317(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	targets := []string{"cmd/gooo#hasPackageOriginQueryOptions", "cmd/gooo#packageOriginQuery", "cmd/gooo#parsePackageExecuteArgs", "cmd/gooo#selectStarterFiles", "cmd/gooo#writePackageOriginInput", "internal/bodyexecution#NativeOriginQuery.hasSubjectSelection", "internal/bodyexecution#ValidateNativeOriginQuery", "internal/bodyexecution#nativeOriginPage.addField", "internal/bodyexecution#nativeOriginPage.fieldFileLocation", "internal/bodyexecution#nativeOriginPage.visitFields", "internal/bodyexecution#selectRecordedNativeOrigins", "internal/bodyexecution#validateOriginSubjectSelection"}
	requireNamedCapacities(t, root, targets)
	t.Log("all accepted-base changed production declarations", len(targets), "original11 retained plus complete predecessor;unchanged75-line helper cap;producer/model0")
}
