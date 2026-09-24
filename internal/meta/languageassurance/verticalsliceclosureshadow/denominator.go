package verticalsliceclosureshadow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	languagesemantic "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic"
	languagesyntax "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	toolchainconformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainconformance"
	toolchainrelease "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainrelease"
	toolchainusecases "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainusecases"
)

func decodeDenominator(raw []byte) (denominator, error) {
	digest := digestBytes(raw)
	if digest != DenominatorDigest && digest != DenominatorMigrationDigest && digest != DenominatorMigrationV23Digest && digest != DenominatorMigrationV24Digest && digest != DenominatorMigrationV25Digest && digest != DenominatorMigrationV26Digest && digest != DenominatorMigrationV27Digest && digest != DenominatorMigrationV28Digest && digest != DenominatorMigrationV29Digest && digest != DenominatorMigrationV30Digest && digest != DenominatorMigrationV31Digest && digest != DenominatorMigrationV32Digest && digest != DenominatorMigrationV33Digest && digest != DenominatorMigrationV34Digest && digest != DenominatorMigrationV35Digest && digest != DenominatorMigrationV36Digest && digest != DenominatorMigrationV37Digest && digest != DenominatorMigrationV38Digest && digest != DenominatorMigrationV39Digest && digest != DenominatorMigrationV40Digest && digest != DenominatorMigrationV41Digest && digest != DenominatorMigrationV42Digest && digest != DenominatorMigrationV43Digest && digest != DenominatorMigrationV44Digest && digest != DenominatorMigrationV45Digest {
		return denominator{}, fmt.Errorf("denominator digest mismatch")
	}
	var value denominator
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return denominator{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return denominator{}, fmt.Errorf("denominator trailing content")
	}
	if err := validateDenominator(value); err != nil {
		return denominator{}, err
	}
	return value, nil
}

func validateDenominator(value denominator) error {
	expected := expectedBoundarySpecsForVersion(value.Version)
	validHeader := (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v21" && value.Version == 21) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v22" && value.Version == 22) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v23" && value.Version == 23) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v24" && value.Version == 24) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v25" && value.Version == 25) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v26" && value.Version == 26) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v27" && value.Version == 27) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v28" && value.Version == 28) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v29" && value.Version == 29) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v30" && value.Version == 30) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v31" && value.Version == 31) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v32" && value.Version == 32) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v33" && value.Version == 33) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v34" && value.Version == 34) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v35" && value.Version == 35) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v36" && value.Version == 36) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v37" && value.Version == 37) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v38" && value.Version == 38) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v39" && value.Version == 39) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v40" && value.Version == 40) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v41" && value.Version == 41) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v42" && value.Version == 42) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v43" && value.Version == 43) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v44" && value.Version == 44) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v45" && value.Version == 45) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v46" && value.Version == 46)
	if value.Schema != "gooo/vertical-slice-boundary-denominator/v1" ||
		!validHeader || len(value.Boundaries) != len(expected) {
		return fmt.Errorf("denominator header mismatch")
	}
	links := 0
	for index, spec := range value.Boundaries {
		if spec != expected[index] {
			return fmt.Errorf("denominator boundary %d mismatch", index)
		}
		links += spec.LinkTarget
	}
	if links != linkTotal {
		return fmt.Errorf("denominator link total mismatch")
	}
	return nil
}

func expectedBoundarySpecsForVersion(version int) []boundarySpec {
	expected := expectedBoundarySpecs()
	switch version {
	case 45:
		expected[0].Target = 77
	case 46:
		expected[0].Target = 79
	}
	return expected
}

func expectedBoundarySpecs() []boundarySpec {
	return []boundarySpec{
		{"syntax", languagesyntax.ReportSchema, "prove-language-syntax-roundtrip", languagesyntax.FixedCapabilityTotal, 1},
		{"semantics", languagesemantic.ReportSchema, "prove-staged-semantic-model", languagesemantic.FixedTotal, 2},
		{"binding", "gooo/language-semantic-readiness-binding/v2", "bind-semantic-readiness-evidence", 12, 2},
		{"use-cases", toolchainusecases.ReportSchema, "execute-versioned-use-cases", 3, 1},
		{"toolchain", toolchainconformance.Schema, toolchainconformance.ExpectedMetaOperation,
			toolchainconformance.ExpectedCaseCount, 3},
		{"release", toolchainrelease.ReportSchema, toolchainrelease.MetaOperation, 20, 3},
	}
}

func codeMetaOperation(id string) string {
	for _, spec := range expectedBoundarySpecs() {
		if spec.ID == id {
			return spec.MetaOperation
		}
	}
	return ""
}