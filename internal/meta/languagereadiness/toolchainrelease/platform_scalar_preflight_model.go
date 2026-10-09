package toolchainrelease

import (
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func validateScalarPreflightModel(c *bodycodegen.RecordModelCompatibility) error {
	if c == nil || c.Schema != "gooo/record-model-compatibility/v1" || c.Status != "READY_FOR_RANKING" ||
		c.Reason != "SOURCE_INPUT_ENCODED" || c.Model.Schema != "gooo/retained-path-model/v1" || !c.Model.Loaded ||
		c.Model.MetadataSHA256 != "3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202" ||
		c.Model.WeightsSHA256 != "76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f" ||
		c.Model.ResidentTensorBytes != 2096 || c.Model.ModelSchema != jointdecision.SharedThreeSchema ||
		c.Model.FeatureVersion != jointdecision.RecordGraphSharedFeatureVersion ||
		c.Model.ArithmeticVersion != jointdecision.SeparateArithmeticVersion {
		return fmt.Errorf("scalar model preflight verified model identity differs")
	}
	return nil
}
