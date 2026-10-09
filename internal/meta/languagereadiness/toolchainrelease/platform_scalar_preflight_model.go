package toolchainrelease

import (
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func validateScalarPreflightModel(c *bodycodegen.RecordModelCompatibility) error {
	if c == nil || c.Schema != "gooo/record-model-compatibility/v1" || c.Status != "READY_FOR_RANKING" ||
		c.Reason != "SOURCE_INPUT_ENCODED" {
		return fmt.Errorf("scalar model preflight verified model identity differs")
	}
	return validateScalarModelIdentity(c.Model)
}

func validateScalarModelIdentity(m bodycodegen.RetainedModelInfo) error {
	if m.Schema != "gooo/retained-path-model/v1" || !m.Loaded ||
		m.MetadataSHA256 != "3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202" ||
		m.WeightsSHA256 != "76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f" ||
		m.ResidentTensorBytes != 2096 || m.ModelSchema != jointdecision.SharedThreeSchema ||
		m.FeatureVersion != jointdecision.RecordGraphSharedFeatureVersion ||
		m.ArithmeticVersion != jointdecision.SeparateArithmeticVersion {
		return fmt.Errorf("own three-choice model identity differs")
	}
	return nil
}
