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
	if !acceptedDenominatorDigestChunk0(digest) {
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
	validHeader := acceptedDenominatorHeaderChunk0(value)
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
	expectedLinks := linkTotal
	if value.Version <= 46 {
		expectedLinks = 12
	}
	if links != expectedLinks {
		return fmt.Errorf("denominator link total mismatch")
	}
	return nil
}

func expectedBoundarySpecsForVersion(version int) []boundarySpec {
	expected := expectedBoundarySpecs()
	if version >= 21 && version <= 46 {
		expected[5].Target = 20
		expected[5].LinkTarget = 3
	}
	if target, known := historicalSyntaxTargetChunk0(version); known {
		expected[0].Target = target
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
		{"release", toolchainrelease.ReportSchema, toolchainrelease.MetaOperation, 26, 4},
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

func historicalSyntaxTargetChunk0(version int) (int, bool) {
	switch version {
	case 45:
		return 77, true
	case 46:
		return 79, true
	case 47:
		return 79, true
	case 48:
		return 80, true
	case 49:
		return 82, true
	case 65:
		return 98, true
	case 66:
		return 99, true
	case 67:
		return 100, true
	case 68:
		return 101, true
	case 69:
		return 102, true
	case 70:
		return 103, true
	case 71:
		return 104, true
	case 72:
		return 108, true
	case 73:
		return 109, true
	case 74:
		return 110, true
	case 75:
		return 111, true
	case 76:
		return 112, true
	case 77:
		return 113, true
	case 78:
		return 114, true
	case 79:
		return 115, true
	case 80:
		return 116, true
	case 81:
		return 117, true
	case 82:
		return 118, true
	case 83:
		return 119, true
	}
	return historicalSyntaxTargetChunk1(version)
}

func historicalSyntaxTargetChunk1(version int) (int, bool) {
	switch version {
	case 84:
		return 120, true
	case 85:
		return 121, true
	case 86:
		return 122, true
	case 87:
		return 123, true
	case 88:
		return 124, true
	case 89:
		return 125, true
	case 90:
		return 126, true
	case 91:
		return 127, true
	case 92:
		return 128, true
	case 93:
		return 129, true
	case 94:
		return 130, true
	case 95:
		return 131, true
	case 96:
		return 132, true
	case 97:
		return 133, true
	case 98:
		return 134, true
	case 99:
		return 135, true
	case 100:
		return 136, true
	case 101:
		return 137, true
	case 102:
		return 138, true
	case 103:
		return 139, true
	case 104:
		return 140, true
	case 105:
		return 141, true
	case 106:
		return 142, true
	case 107:
		return 143, true
	}
	return historicalSyntaxTargetChunk2(version)
}

func historicalSyntaxTargetChunk2(version int) (int, bool) {
	switch version {
	case 108:
		return 144, true
	case 109:
		return 145, true
	case 110:
		return 146, true
	}
	return 0, false
}

func acceptedDenominatorHeaderChunk0(value denominator) bool {
	switch value.Version {
	case 21:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v21"
	case 22:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v22"
	case 23:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v23"
	case 24:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v24"
	case 25:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v25"
	case 26:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v26"
	case 27:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v27"
	case 28:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v28"
	case 29:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v29"
	case 30:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v30"
	case 31:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v31"
	case 32:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v32"
	case 33:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v33"
	case 34:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v34"
	case 35:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v35"
	case 36:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v36"
	}
	return acceptedDenominatorHeaderChunk1(value)
}

func acceptedDenominatorHeaderChunk1(value denominator) bool {
	switch value.Version {
	case 37:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v37"
	case 38:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v38"
	case 39:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v39"
	case 40:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v40"
	case 41:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v41"
	case 42:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v42"
	case 43:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v43"
	case 44:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v44"
	case 45:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v45"
	case 46:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v46"
	case 47:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v47"
	case 48:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v48"
	case 49:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v49"
	case 50:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v50"
	case 51:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v51"
	case 52:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v52"
	}
	return acceptedDenominatorHeaderChunk2(value)
}

func acceptedDenominatorHeaderChunk2(value denominator) bool {
	switch value.Version {
	case 53:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v53"
	case 54:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v54"
	case 55:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v55"
	case 56:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v56"
	case 57:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v57"
	case 58:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v58"
	case 59:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v59"
	case 60:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v60"
	case 61:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v61"
	case 62:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v62"
	case 63:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v63"
	case 64:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v64"
	case 65:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v65"
	case 66:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v66"
	case 67:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v67"
	case 68:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v68"
	}
	return acceptedDenominatorHeaderChunk3(value)
}

func acceptedDenominatorHeaderChunk3(value denominator) bool {
	switch value.Version {
	case 69:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v69"
	case 70:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v70"
	case 71:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v71"
	case 72:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v72"
	case 73:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v73"
	case 74:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v74"
	case 75:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v75"
	case 76:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v76"
	case 77:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v77"
	case 78:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v78"
	case 79:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v79"
	case 80:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v80"
	case 81:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v81"
	case 82:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v82"
	case 83:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v83"
	case 84:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v84"
	}
	return acceptedDenominatorHeaderChunk4(value)
}

func acceptedDenominatorHeaderChunk4(value denominator) bool {
	switch value.Version {
	case 85:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v85"
	case 86:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v86"
	case 87:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v87"
	case 88:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v88"
	case 89:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v89"
	case 90:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v90"
	case 91:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v91"
	case 92:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v92"
	case 93:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v93"
	case 94:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v94"
	case 95:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v95"
	case 96:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v96"
	case 97:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v97"
	case 98:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v98"
	case 99:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v99"
	case 100:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v100"
	}
	return acceptedDenominatorHeaderChunk5(value)
}

func acceptedDenominatorHeaderChunk5(value denominator) bool {
	switch value.Version {
	case 101:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v101"
	case 102:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v102"
	case 103:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v103"
	case 104:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v104"
	case 105:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v105"
	case 106:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v106"
	case 107:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v107"
	case 108:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v108"
	case 109:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v109"
	case 110:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v110"
	case 111:
		return value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v111"
	}
	return false
}

func acceptedDenominatorDigestChunk0(digest string) bool {
	switch digest {
	case DenominatorDigest:
		return true
	case DenominatorMigrationDigest:
		return true
	case DenominatorMigrationV23Digest:
		return true
	case DenominatorMigrationV24Digest:
		return true
	case DenominatorMigrationV25Digest:
		return true
	case DenominatorMigrationV26Digest:
		return true
	case DenominatorMigrationV27Digest:
		return true
	case DenominatorMigrationV28Digest:
		return true
	case DenominatorMigrationV29Digest:
		return true
	case DenominatorMigrationV30Digest:
		return true
	case DenominatorMigrationV31Digest:
		return true
	case DenominatorMigrationV32Digest:
		return true
	case DenominatorMigrationV33Digest:
		return true
	case DenominatorMigrationV34Digest:
		return true
	case DenominatorMigrationV35Digest:
		return true
	case DenominatorMigrationV36Digest:
		return true
	}
	return acceptedDenominatorDigestChunk1(digest)
}

func acceptedDenominatorDigestChunk1(digest string) bool {
	switch digest {
	case DenominatorMigrationV37Digest:
		return true
	case DenominatorMigrationV38Digest:
		return true
	case DenominatorMigrationV39Digest:
		return true
	case DenominatorMigrationV40Digest:
		return true
	case DenominatorMigrationV41Digest:
		return true
	case DenominatorMigrationV42Digest:
		return true
	case DenominatorMigrationV43Digest:
		return true
	case DenominatorMigrationV44Digest:
		return true
	case DenominatorMigrationV45Digest:
		return true
	case DenominatorMigrationV46Digest:
		return true
	case DenominatorMigrationV47Digest:
		return true
	case DenominatorMigrationV48Digest:
		return true
	case DenominatorMigrationV49Digest:
		return true
	case DenominatorMigrationV50Digest:
		return true
	case DenominatorMigrationV51Digest:
		return true
	case DenominatorMigrationV52Digest:
		return true
	}
	return acceptedDenominatorDigestChunk2(digest)
}

func acceptedDenominatorDigestChunk2(digest string) bool {
	switch digest {
	case DenominatorMigrationV53Digest:
		return true
	case DenominatorMigrationV54Digest:
		return true
	case DenominatorMigrationV55Digest:
		return true
	case DenominatorMigrationV56Digest:
		return true
	case DenominatorMigrationV57Digest:
		return true
	case DenominatorMigrationV58Digest:
		return true
	case DenominatorMigrationV59Digest:
		return true
	case DenominatorMigrationV60Digest:
		return true
	case DenominatorMigrationV61Digest:
		return true
	case DenominatorMigrationV62Digest:
		return true
	case DenominatorMigrationV63Digest:
		return true
	case DenominatorMigrationV64Digest:
		return true
	case DenominatorMigrationV65Digest:
		return true
	case DenominatorMigrationV66Digest:
		return true
	case DenominatorMigrationV67Digest:
		return true
	case DenominatorMigrationV68Digest:
		return true
	}
	return acceptedDenominatorDigestChunk3(digest)
}

func acceptedDenominatorDigestChunk3(digest string) bool {
	switch digest {
	case DenominatorMigrationV69Digest:
		return true
	case DenominatorMigrationV70Digest:
		return true
	case DenominatorMigrationV71Digest:
		return true
	case DenominatorMigrationV72Digest:
		return true
	case DenominatorMigrationV73Digest:
		return true
	case DenominatorMigrationV74Digest:
		return true
	case DenominatorMigrationV75Digest:
		return true
	case DenominatorMigrationV76Digest:
		return true
	case DenominatorMigrationV77Digest:
		return true
	case DenominatorMigrationV78Digest:
		return true
	case DenominatorMigrationV79Digest:
		return true
	case DenominatorMigrationV80Digest:
		return true
	case DenominatorMigrationV81Digest:
		return true
	case DenominatorMigrationV82Digest:
		return true
	case DenominatorMigrationV83Digest:
		return true
	case DenominatorMigrationV84Digest:
		return true
	}
	return acceptedDenominatorDigestChunk4(digest)
}

func acceptedDenominatorDigestChunk4(digest string) bool {
	switch digest {
	case DenominatorMigrationV85Digest:
		return true
	case DenominatorMigrationV86Digest:
		return true
	case DenominatorMigrationV87Digest:
		return true
	case DenominatorMigrationV88Digest:
		return true
	case DenominatorMigrationV89Digest:
		return true
	case DenominatorMigrationV90Digest:
		return true
	case DenominatorMigrationV91Digest:
		return true
	case DenominatorMigrationV92Digest:
		return true
	case DenominatorMigrationV93Digest:
		return true
	case DenominatorMigrationV94Digest:
		return true
	case DenominatorMigrationV95Digest:
		return true
	case DenominatorMigrationV96Digest:
		return true
	case DenominatorMigrationV97Digest:
		return true
	case DenominatorMigrationV98Digest:
		return true
	case DenominatorMigrationV99Digest:
		return true
	case DenominatorMigrationV100Digest:
		return true
	}
	return acceptedDenominatorDigestChunk5(digest)
}

func acceptedDenominatorDigestChunk5(digest string) bool {
	switch digest {
	case DenominatorMigrationV101Digest:
		return true
	case DenominatorMigrationV102Digest:
		return true
	case DenominatorMigrationV103Digest:
		return true
	case DenominatorMigrationV104Digest:
		return true
	case DenominatorMigrationV105Digest:
		return true
	case DenominatorMigrationV106Digest:
		return true
	case DenominatorMigrationV107Digest:
		return true
	case DenominatorMigrationV108Digest:
		return true
	case DenominatorMigrationV109Digest:
		return true
	case DenominatorMigrationV110Digest:
		return true
	case DenominatorMigrationV111Digest:
		return true
	}
	return false
}
