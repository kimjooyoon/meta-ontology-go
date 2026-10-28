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
	if ((((((((((((((((((((((((((((((((((((((((((((((((((((((((((digest != DenominatorDigest && digest != DenominatorMigrationDigest && digest != DenominatorMigrationV23Digest && digest != DenominatorMigrationV24Digest && digest != DenominatorMigrationV25Digest && digest != DenominatorMigrationV26Digest && digest != DenominatorMigrationV27Digest && digest != DenominatorMigrationV28Digest && digest != DenominatorMigrationV29Digest && digest != DenominatorMigrationV30Digest && digest != DenominatorMigrationV31Digest && digest != DenominatorMigrationV32Digest && digest != DenominatorMigrationV33Digest && digest != DenominatorMigrationV34Digest && digest != DenominatorMigrationV35Digest && digest != DenominatorMigrationV36Digest && digest != DenominatorMigrationV37Digest && digest != DenominatorMigrationV38Digest && digest != DenominatorMigrationV39Digest && digest != DenominatorMigrationV40Digest && digest != DenominatorMigrationV41Digest && digest != DenominatorMigrationV42Digest && digest != DenominatorMigrationV43Digest && digest != DenominatorMigrationV44Digest && digest != DenominatorMigrationV45Digest && digest != DenominatorMigrationV46Digest && digest != DenominatorMigrationV47Digest && digest != DenominatorMigrationV48Digest && digest != DenominatorMigrationV49Digest) && digest != DenominatorMigrationV50Digest) && digest != DenominatorMigrationV51Digest) && digest != DenominatorMigrationV52Digest) && digest != DenominatorMigrationV53Digest) && digest != DenominatorMigrationV54Digest) && digest != DenominatorMigrationV55Digest) && digest != DenominatorMigrationV56Digest) && digest != DenominatorMigrationV57Digest) && digest != DenominatorMigrationV58Digest) && digest != DenominatorMigrationV59Digest) && digest != DenominatorMigrationV60Digest) && digest != DenominatorMigrationV61Digest) && digest != DenominatorMigrationV62Digest) && digest != DenominatorMigrationV63Digest) && digest != DenominatorMigrationV64Digest) && digest != DenominatorMigrationV65Digest) && digest != DenominatorMigrationV66Digest) && digest != DenominatorMigrationV67Digest) && digest != DenominatorMigrationV68Digest) && digest != DenominatorMigrationV69Digest) && digest != DenominatorMigrationV70Digest) && digest != DenominatorMigrationV71Digest && digest != DenominatorMigrationV72Digest) && digest != DenominatorMigrationV73Digest) && digest != DenominatorMigrationV74Digest) && digest != DenominatorMigrationV75Digest) && digest != DenominatorMigrationV76Digest) && digest != DenominatorMigrationV77Digest) && digest != DenominatorMigrationV78Digest) && digest != DenominatorMigrationV79Digest) && digest != DenominatorMigrationV80Digest) && digest != DenominatorMigrationV81Digest) && digest != DenominatorMigrationV82Digest) && digest != DenominatorMigrationV83Digest) && digest != DenominatorMigrationV84Digest) && digest != DenominatorMigrationV85Digest) && digest != DenominatorMigrationV86Digest) && digest != DenominatorMigrationV87Digest) && digest != DenominatorMigrationV88Digest) && digest != DenominatorMigrationV89Digest) && digest != DenominatorMigrationV90Digest) && digest != DenominatorMigrationV91Digest) && digest != DenominatorMigrationV92Digest) && digest != DenominatorMigrationV93Digest) && digest != DenominatorMigrationV94Digest) && digest != DenominatorMigrationV95Digest) && digest != DenominatorMigrationV96Digest) && digest != DenominatorMigrationV97Digest) && digest != DenominatorMigrationV98Digest) && digest != DenominatorMigrationV99Digest) && digest != DenominatorMigrationV100Digest) && digest != DenominatorMigrationV101Digest) && digest != DenominatorMigrationV102Digest) && digest != DenominatorMigrationV103Digest) && digest != DenominatorMigrationV104Digest) && digest != DenominatorMigrationV105Digest) && digest != DenominatorMigrationV106Digest) && digest != DenominatorMigrationV107Digest) && digest != DenominatorMigrationV108Digest {
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
	validHeader := (((((((((((((((((((((((((((((((((((((((((((((((((((((((((((value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v21" && value.Version == 21) ||
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
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v46" && value.Version == 46) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v47" && value.Version == 47) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v48" && value.Version == 48) ||
		(value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v49" && value.Version == 49)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v50" && value.Version == 50)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v51" && value.Version == 51)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v52" && value.Version == 52)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v53" && value.Version == 53)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v54" && value.Version == 54)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v55" && value.Version == 55)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v56" && value.Version == 56)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v57" && value.Version == 57)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v58" && value.Version == 58)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v59" && value.Version == 59)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v60" && value.Version == 60)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v61" && value.Version == 61)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v62" && value.Version == 62)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v63" && value.Version == 63)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v64" && value.Version == 64)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v65" && value.Version == 65)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v66" && value.Version == 66)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v67" && value.Version == 67)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v68" && value.Version == 68)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v69" && value.Version == 69)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v70" && value.Version == 70)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v71" && value.Version == 71) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v72" && value.Version == 72)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v73" && value.Version == 73)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v74" && value.Version == 74)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v75" && value.Version == 75)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v76" && value.Version == 76)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v77" && value.Version == 77)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v78" && value.Version == 78)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v79" && value.Version == 79)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v80" && value.Version == 80)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v81" && value.Version == 81)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v82" && value.Version == 82)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v83" && value.Version == 83)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v84" && value.Version == 84)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v85" && value.Version == 85)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v86" && value.Version == 86)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v87" && value.Version == 87)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v88" && value.Version == 88)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v89" && value.Version == 89)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v90" && value.Version == 90)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v91" && value.Version == 91)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v92" && value.Version == 92)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v93" && value.Version == 93)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v94" && value.Version == 94)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v95" && value.Version == 95)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v96" && value.Version == 96)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v97" && value.Version == 97)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v98" && value.Version == 98)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v99" && value.Version == 99)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v100" && value.Version == 100)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v101" && value.Version == 101)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v102" && value.Version == 102)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v103" && value.Version == 103)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v104" && value.Version == 104)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v105" && value.Version == 105)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v106" && value.Version == 106)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v107" && value.Version == 107)) || (value.DenominatorID == "gooo.denominator.capability.vertical-slice-closure.v108" && value.Version == 108)
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
	return 0, false
}
