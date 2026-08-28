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
	if digestBytes(raw) != DenominatorDigest {
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
	expected := expectedBoundarySpecs()
	if value.Schema != "gooo/vertical-slice-boundary-denominator/v1" ||
		value.DenominatorID != "gooo.denominator.capability.vertical-slice-closure.v19" ||
		value.Version != 19 || len(value.Boundaries) != len(expected) {
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

func expectedBoundarySpecs() []boundarySpec {
	return []boundarySpec{
		{"syntax", languagesyntax.ReportSchema, "prove-language-syntax-roundtrip", languagesyntax.FixedTotal, 1},
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
