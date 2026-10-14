package extractor

import (
	"os"
	"testing"
)

func TestDerivedCandidateCapacity27593(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../../../.."
	}
	targets := []string{"cmd/gooo#hasPackageOriginQueryOptions", "cmd/gooo#packageOriginQuery", "cmd/gooo#parsePackageExecuteArgs", "cmd/gooo#runBodyOrigins", "cmd/gooo#selectStarterFiles", "cmd/gooo#writeBodyDerivedCandidateOrigins", "cmd/gooo#writeBodyFillCandidateOrigins", "cmd/gooo#writeBodyFillOriginCandidates", "cmd/gooo#writePackageOriginInput", "internal/bodycodegen#InspectSourceFillCandidateOrigins", "internal/bodycodegen#InspectSourceFillDerivedOrigins", "internal/bodycodegen#InspectSourceFillOriginCandidates", "internal/bodycodegen#IsSourceFillDerivation", "internal/bodycodegen#SourceFillCandidateOriginInputs", "internal/bodycodegen#SourceFillDerivedOriginInputs", "internal/bodycodegen#sourceFillCandidateOriginRow", "internal/bodycodegen#sourceFillDerivedOriginRow", "internal/bodyexecution#NativeOriginQuery.hasSubjectSelection", "internal/bodyexecution#ValidateNativeOriginQuery", "internal/bodyexecution#nativeOriginPage.addField", "internal/bodyexecution#nativeOriginPage.fieldFileLocation", "internal/bodyexecution#nativeOriginPage.visitFields", "internal/bodyexecution#selectRecordedNativeOrigins", "internal/bodyexecution#validateOriginSubjectSelection", "internal/meta/languageassurance/verticalsliceclosureshadow#activeDenominatorChunk2", "internal/meta/languageassurance/verticalsliceclosureshadow#activeDenominatorDigestChunk2", "internal/meta/languageassurance/verticalsliceclosureshadow#decodeDenominator", "internal/meta/languageassurance/verticalsliceclosureshadow#expectedBoundarySpecsForVersion", "internal/meta/languageassurance/verticalsliceclosureshadow#validateDenominator", "internal/meta/languagereadiness/languagesyntax#expectedRegistry", "internal/meta/languagereadiness/languagesyntax#expectedSyntaxCaseGroup0", "internal/meta/languagereadiness/languagesyntax#expectedSyntaxCaseGroup1", "internal/meta/languagereadiness/languagesyntax#expectedSyntaxCaseGroup2", "internal/meta/languagereadiness/languagesyntax#expectedSyntaxCaseGroup3", "internal/meta/languagereadiness/languagesyntax#expectedSyntaxCaseGroup4", "internal/meta/languagereadiness/languagesyntax#expectedSyntaxCaseGroup5", "internal/meta/languagereadiness/languagesyntax#expectedSyntaxCases", "internal/meta/languagereadiness/languagesyntax#expectedSyntaxPackageUnits"}
	requireNamedCapacities(t, root, targets)
	t.Log("derived27593:all38 changed production declarations;prior17 retained plus grammar/list/registry/denominator changes;unchanged75-line rendered helper cap;producer/model0")
}
